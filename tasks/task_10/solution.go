package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Clock interface {
	Now() time.Time
}

type Task struct {
	ID        string
	Title     string
	Done      bool
	UpdatedAt time.Time
}

type TaskRepo interface {
	Create(title string) (Task, error)
	Get(id string) (Task, bool)
	List() []Task
	SetDone(id string, done bool) (Task, error)
}

type InMemoryTaskRepo struct {
	mu      sync.RWMutex
	tasks   map[string]Task
	clock   Clock
	counter atomic.Int64
}

func NewInMemoryTaskRepo(clock Clock) *InMemoryTaskRepo {
	return &InMemoryTaskRepo{
		tasks: make(map[string]Task),
		clock: clock,
	}
}

func (r *InMemoryTaskRepo) nextID() string {
	return fmt.Sprintf("%016d", r.counter.Add(1))
}

func (r *InMemoryTaskRepo) Create(title string) (Task, error) {
	task := Task{
		ID:        r.nextID(),
		Title:     title,
		Done:      false,
		UpdatedAt: r.clock.Now(),
	}
	r.mu.Lock()
	r.tasks[task.ID] = task
	r.mu.Unlock()
	return task, nil
}

func (r *InMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mu.RLock()
	task, ok := r.tasks[id]
	r.mu.RUnlock()
	return task, ok
}

func (r *InMemoryTaskRepo) List() []Task {
	r.mu.RLock()
	result := make([]Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		result = append(result, t)
	}
	r.mu.RUnlock()
	return result
}

func (r *InMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	task, ok := r.tasks[id]
	if !ok {
		return Task{}, errors.New("task not found: " + id)
	}
	task.Done = done
	task.UpdatedAt = r.clock.Now()
	r.tasks[id] = task
	return task, nil
}

type taskJSON struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func toJSON(t Task) taskJSON {
	return taskJSON{ID: t.ID, Title: t.Title, Done: t.Done, UpdatedAt: t.UpdatedAt}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func NewHTTPHandler(repo TaskRepo) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /tasks", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Title string `json:"title"`
		}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.Title) == "" {
			http.Error(w, "title must not be empty", http.StatusBadRequest)
			return
		}
		task, _ := repo.Create(req.Title)
		writeJSON(w, http.StatusCreated, toJSON(task))
	})

	mux.HandleFunc("GET /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		task, ok := repo.Get(id)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, toJSON(task))
	})

	mux.HandleFunc("GET /tasks", func(w http.ResponseWriter, r *http.Request) {
		tasks := repo.List()
		sort.Slice(tasks, func(i, j int) bool {
			if tasks[i].UpdatedAt.Equal(tasks[j].UpdatedAt) {
				return tasks[i].ID < tasks[j].ID
			}
			return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
		})
		result := make([]taskJSON, len(tasks))
		for i, t := range tasks {
			result[i] = toJSON(t)
		}
		writeJSON(w, http.StatusOK, result)
	})

	mux.HandleFunc("PATCH /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var req struct {
			Done *bool `json:"done"`
		}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Done == nil {
			http.Error(w, "done is required", http.StatusBadRequest)
			return
		}
		task, err := repo.SetDone(id, *req.Done)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, toJSON(task))
	})

	return mux
}
