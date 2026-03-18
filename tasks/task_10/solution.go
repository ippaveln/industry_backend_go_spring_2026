package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type TaskRepo interface {
	Create(title string) (Task, error)
	Get(id string) (Task, bool)
	List() []Task
	SetDone(id string, done bool) (Task, error)
}

type Task struct {
	ID        string
	Title     string
	Done      bool
	UpdatedAt time.Time
}

type InMemoryTaskVault struct {
	mu        sync.RWMutex
	tasks     map[string]Task
	clock     Clock
	currNewId uint
}

func (i *InMemoryTaskVault) Create(title string) (Task, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	task := Task{ID: strconv.Itoa(int(i.currNewId)), Title: title, UpdatedAt: i.clock.Now()}
	i.tasks[strconv.Itoa(int(i.currNewId))] = task
	i.currNewId++
	return task, nil
}

func (i *InMemoryTaskVault) Get(id string) (Task, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	task, ok := i.tasks[id]
	return task, ok
}

func (i *InMemoryTaskVault) List() []Task {
	i.mu.RLock()
	defer i.mu.RUnlock()
	tasks := make([]Task, 0, len(i.tasks))
	for _, task := range i.tasks {
		tasks = append(tasks, task)
	}
	sort.Slice(tasks, func(a, b int) bool {
		ta, tb := tasks[a].UpdatedAt, tasks[b].UpdatedAt
		if !ta.Equal(tb) {
			return ta.After(tb)
		}
		return tasks[a].ID < tasks[b].ID
	})
	return tasks
}

func (i *InMemoryTaskVault) SetDone(id string, done bool) (Task, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	task, ok := i.tasks[id]
	if !ok {
		return task, errors.New("task not found")
	}
	task.Done = done
	task.UpdatedAt = i.clock.Now()
	i.tasks[id] = task
	return task, nil
}

func NewInMemoryTaskRepo(clock Clock) TaskRepo {
	return &InMemoryTaskVault{tasks: make(map[string]Task, 8), clock: clock}
}

func NewHTTPHandler(taskRepo TaskRepo) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /tasks", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Title *string `json:"title"`
		}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if req.Title == nil {
			http.Error(w, "empty JSON", http.StatusBadRequest)
		}
		*req.Title = strings.TrimSpace(*req.Title)
		if *req.Title == "" {
			http.Error(w, "title is empty", http.StatusBadRequest)
			return
		}
		task, err := taskRepo.Create(*req.Title)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Формирование ответа в правильном формате (camelCase для JSON)
		resp := map[string]interface{}{
			"id":        task.ID,
			"title":     task.Title,
			"done":      task.Done,
			"updatedAt": task.UpdatedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(resp); err != nil {

		}
	})

	mux.HandleFunc("GET /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		task, ok := taskRepo.Get(id)
		if !ok {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}

		resp := map[string]any{
			"id":        task.ID,
			"title":     task.Title,
			"done":      task.Done,
			"updatedAt": task.UpdatedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {

		}
	})

	mux.HandleFunc("GET /tasks", func(w http.ResponseWriter, r *http.Request) {
		tasks := taskRepo.List()

		resp := make([]map[string]any, len(tasks))
		for i, task := range tasks {
			resp[i] = map[string]any{
				"id":        task.ID,
				"title":     task.Title,
				"done":      task.Done,
				"updatedAt": task.UpdatedAt,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {

		}
	})

	mux.HandleFunc("PATCH /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Done *bool `json:"done"`
		}
		id := r.PathValue("id")
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if req.Done == nil {
			http.Error(w, "empty JSON", http.StatusBadRequest)
			return
		}

		task, err := taskRepo.SetDone(id, *req.Done)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		resp := map[string]any{
			"id":        task.ID,
			"title":     task.Title,
			"done":      task.Done,
			"updatedAt": task.UpdatedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {

		}
	})

	return mux
}
