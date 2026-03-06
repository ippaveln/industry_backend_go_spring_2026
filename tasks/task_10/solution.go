package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

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

type InMemoryTaskRepo struct {
	data      map[string]Task
	clock     Clock
	idCounter int
	mu        sync.RWMutex
}

func NewInMemoryTaskRepo(c Clock) *InMemoryTaskRepo {
	return &InMemoryTaskRepo{
		data:  make(map[string]Task),
		clock: c,
	}
}

func (r *InMemoryTaskRepo) Create(title string) (Task, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return Task{}, errors.New("title can't be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.idCounter++
	id := strconv.Itoa(r.idCounter)

	task := Task{
		ID:        id,
		Title:     title,
		Done:      false,
		UpdatedAt: r.clock.Now(),
	}

	r.data[id] = task

	return task, nil
}

func (r *InMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.data[id]
	return task, ok
}

func (r *InMemoryTaskRepo) List() []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]Task, 0, len(r.data))
	for _, v := range r.data {
		tasks = append(tasks, v)
	}

	slices.SortFunc(tasks, func(a, b Task) int {
		if a.UpdatedAt.After(b.UpdatedAt) {
			return -1
		}
		if a.UpdatedAt.Before(b.UpdatedAt) {
			return 1
		}
		idA, _ := strconv.Atoi(a.ID)
		idB, _ := strconv.Atoi(b.ID)

		return idA - idB
	})
	return tasks
}

func (r *InMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.data[id]
	if !ok {
		return Task{}, fmt.Errorf("task not found: id: %s", id)
	}
	task.Done = done
	task.UpdatedAt = r.clock.Now()
	r.data[id] = task

	return task, nil
}

type taskHandler struct {
	repo TaskRepo
}

func NewHTTPHandler(repo TaskRepo) http.Handler {
	mux := http.NewServeMux()
	h := &taskHandler{repo: repo}

	mux.HandleFunc("POST /tasks", h.create)
	mux.HandleFunc("GET /tasks/{id}", h.get)
	mux.HandleFunc("GET /tasks", h.list)
	mux.HandleFunc("PATCH /tasks/{id}", h.setdone)

	return mux
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *taskHandler) create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task, err := h.repo.Create(req.Title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (h *taskHandler) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	task, ok := h.repo.Get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *taskHandler) list(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.repo.List())
}

func (h *taskHandler) setdone(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, `"done" field is required`, http.StatusBadRequest)
		return
	}

	task, err := h.repo.SetDone(id, *req.Done)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, task)
}
