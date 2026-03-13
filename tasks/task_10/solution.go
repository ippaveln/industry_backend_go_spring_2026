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
	ID        string
	Title     string
	Done      bool
	UpdatedAt time.Time
}

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
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
	counter int64
}

func NewInMemoryTaskRepo(clock Clock) *InMemoryTaskRepo {
	return &InMemoryTaskRepo{
		tasks: make(map[string]Task),
		clock: clock,
	}
}

func (r *InMemoryTaskRepo) Create(title string) (Task, error) {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return Task{}, fmt.Errorf("title cannot be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counter++
	id := fmt.Sprintf("%d", r.counter)
	now := r.clock.Now()
	task := Task{
		ID:        id,
		Title:     trimmed,
		Done:      false,
		UpdatedAt: now,
	}
	r.tasks[id] = task
	return task, nil
}

func (r *InMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	task, ok := r.tasks[id]
	return task, ok
}

func (r *InMemoryTaskRepo) List() []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tasks := make([]Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		tasks = append(tasks, t)
	}
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].UpdatedAt.Equal(tasks[j].UpdatedAt) {
			return tasks[i].ID < tasks[j].ID
		}
		return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
	})
	return tasks
}

func (r *InMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	task, ok := r.tasks[id]
	if !ok {
		return Task{}, fmt.Errorf("task with id %s not found", id)
	}
	task.Done = done
	task.UpdatedAt = r.clock.Now()
	r.tasks[id] = task
	return task, nil
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
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.Title) == "" {
			http.Error(w, "title cannot be empty", http.StatusBadRequest)
			return
		}
		task, err := repo.Create(req.Title)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
	})

	mux.HandleFunc("GET /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}
		task, ok := repo.Get(id)
		if !ok {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	})

	mux.HandleFunc("GET /tasks", func(w http.ResponseWriter, r *http.Request) {
		tasks := repo.List()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
	})

	mux.HandleFunc("PATCH /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}
		var req struct {
			Done *bool `json:"done"`
		}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if req.Done == nil {
			http.Error(w, "missing done field", http.StatusBadRequest)
			return
		}
		task, err := repo.SetDone(id, *req.Done)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	})

	return mux
}
