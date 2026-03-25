package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
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
	mu        sync.RWMutex
	tasks     map[string]Task
	idCounter uint64
	clock     Clock
}

func NewInMemoryTaskRepo(clock Clock) *InMemoryTaskRepo {
	return &InMemoryTaskRepo{
		tasks: make(map[string]Task),
		clock: clock,
	}
}

func (r *InMemoryTaskRepo) Create(title string) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, errors.New("title cannot be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	r.idCounter++
	id := fmt.Sprintf("%d", r.idCounter)
	task := Task{
		ID:        id,
		Title:     title,
		Done:      false,
		UpdatedAt: r.clock.Now(),
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
	tasks := make([]Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		tasks = append(tasks, t)
	}
	r.mu.RUnlock()

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
		return Task{}, errors.New("task not found")
	}
	task.Done = done
	task.UpdatedAt = r.clock.Now()
	r.tasks[id] = task
	return task, nil
}

type HTTPHandler struct {
	repo TaskRepo
}

func NewHTTPHandler(repo TaskRepo) *HTTPHandler {
	return &HTTPHandler{repo: repo}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] != "tasks" {
		http.NotFound(w, r)
		return
	}

	// POST /tasks
	if r.Method == http.MethodPost && len(parts) == 1 {
		h.createTask(w, r)
		return
	}

	// GET /tasks/{id}
	if r.Method == http.MethodGet && len(parts) == 2 && parts[1] != "" {
		h.getTask(w, r, parts[1])
		return
	}

	// GET /tasks
	if r.Method == http.MethodGet && len(parts) == 1 {
		h.listTasks(w, r)
		return
	}

	// PATCH /tasks/{id}
	if r.Method == http.MethodPatch && len(parts) == 2 && parts[1] != "" {
		h.patchTask(w, r, parts[1])
		return
	}

	http.NotFound(w, r)
}

func (h *HTTPHandler) createTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		http.Error(w, "title cannot be empty", http.StatusBadRequest)
		return
	}
	task, err := h.repo.Create(req.Title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *HTTPHandler) getTask(w http.ResponseWriter, r *http.Request, id string) {
	task, ok := h.repo.Get(id)
	if !ok {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(task)
}

func (h *HTTPHandler) listTasks(w http.ResponseWriter, r *http.Request) {
	tasks := h.repo.List()
	json.NewEncoder(w).Encode(tasks)
}

func (h *HTTPHandler) patchTask(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		Done *bool `json:"done"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Done == nil {
		http.Error(w, "missing value for field 'done'", http.StatusBadRequest)
		return
	}
	task, err := h.repo.SetDone(id, *req.Done)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(task)
}
