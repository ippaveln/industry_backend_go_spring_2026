package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TaskRepo interface {
	Create(title string) (Task, error)
	Get(id string) (Task, bool)
	List() []Task
	SetDone(id string, done bool) (Task, error)
}

type Clock interface {
	Now() time.Time
}

type InMemoryTaskRepo struct {
	mtx     sync.RWMutex
	tasks   map[string]Task
	clock   Clock
	counter int
}

func NewInMemoryTaskRepo(clock Clock) *InMemoryTaskRepo {
	return &InMemoryTaskRepo{
		tasks: make(map[string]Task),
		clock: clock,
	}
}

func (r *InMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	task, ok := r.tasks[id]
	return task, ok
}

func (r *InMemoryTaskRepo) Create(title string) (Task, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.counter++

	id := fmt.Sprintf("%d", r.counter)

	task := Task{
		ID:        id,
		Title:     title,
		UpdatedAt: r.clock.Now(),
		Done:      false,
	}

	r.tasks[id] = task
	return task, nil
}

func (r *InMemoryTaskRepo) List() []Task {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	result := make([]Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		result = append(result, t)
	}

	sort.Slice(result, func(i, j int) bool {
		if !result[i].UpdatedAt.Equal(result[j].UpdatedAt) {
			return result[i].UpdatedAt.After(result[j].UpdatedAt)
		}
		return result[i].ID < result[j].ID
	})

	return result
}

func (r *InMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	task, ok := r.tasks[id]
	if !ok {
		return Task{}, fmt.Errorf("task not found")
	}

	task.Done = done
	task.UpdatedAt = r.clock.Now()
	r.tasks[id] = task
	return task, nil
}

// -----------------------------------------------------

type HTTPHandler struct {
	repo TaskRepo
}

func NewHTTPHandler(repo TaskRepo) *HTTPHandler {
	return &HTTPHandler{repo: repo}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/tasks" && r.Method == http.MethodPost {
		h.createTask(w, r)
		return
	}

	if path == "/tasks" && r.Method == http.MethodGet {
		h.listTasks(w, r)
		return
	}

	if after, ok :=strings.CutPrefix(path, "/tasks/"); ok  {
		id := after
		if id == "" {
			http.Error(w, "id param is required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.getTask(w, r, id)
		case http.MethodPatch:
			h.updateTask(w, r, id)
		default:
			http.NotFound(w, r)
		}
		return
	}

	http.NotFound(w, r)
}

func (h *HTTPHandler) createTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid jsonn", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	task, err := h.repo.Create(title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Необязательно для тестов, но правильно будет добавить:
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *HTTPHandler) getTask(w http.ResponseWriter, r *http.Request, id string) {
	task, ok := h.repo.Get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *HTTPHandler) listTasks(w http.ResponseWriter, r *http.Request) {
	tasks := h.repo.List()
	if tasks == nil {
		tasks = []Task{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *HTTPHandler) updateTask(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		Done *bool `json:"done"`
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Done == nil {
		http.Error(w, "done field is required", http.StatusBadRequest)
		return
	}

	task, err := h.repo.SetDone(id, *req.Done)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}
