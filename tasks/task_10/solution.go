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
	mu    sync.RWMutex
	tasks map[string]Task
	clock Clock
	idSeq int64
}

func NewInMemoryTaskRepo(clock Clock) *InMemoryTaskRepo {
	return &InMemoryTaskRepo{
		tasks: make(map[string]Task),
		clock: clock,
	}
}

func (r *InMemoryTaskRepo) Create(title string) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.idSeq++
	id := strconv.FormatInt(r.idSeq, 10)
	now := r.clock.Now()
	task := Task{
		ID:        id,
		Title:     title,
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
	list := make([]Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		list = append(list, t)
	}
	return list
}

func (r *InMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	task, ok := r.tasks[id]
	if !ok {
		return Task{}, errors.New("not found")
	}
	task.Done = done
	task.UpdatedAt = r.clock.Now()
	r.tasks[id] = task
	return task, nil
}

type TaskDTO struct {
	ID        string
	Title     string
	Done      bool
	UpdatedAt time.Time
}

func toDTO(t Task) TaskDTO {
	return TaskDTO{
		ID:        t.ID,
		Title:     t.Title,
		Done:      t.Done,
		UpdatedAt: t.UpdatedAt,
	}
}

type handler struct {
	repo TaskRepo
}

func NewHTTPHandler(repo TaskRepo) http.Handler {
	return &handler{repo: repo}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/tasks" {
		switch r.Method {
		case http.MethodPost:
			h.handleCreate(w, r)
			return
		case http.MethodGet:
			h.handleList(w)
			return
		}
	} else if id, ok := strings.CutPrefix(path, "/tasks/"); ok {
		switch r.Method {
		case http.MethodGet:
			h.handleGet(w, id)
			return
		case http.MethodPatch:
			h.handlePatch(w, r, id)
			return
		}
	}
	http.NotFound(w, r)
}

func (h *handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		http.Error(w, "empty title", http.StatusBadRequest)
		return
	}
	task, err := h.repo.Create(title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toDTO(task))
}

func (h *handler) handleGet(w http.ResponseWriter, id string) {
	task, ok := h.repo.Get(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toDTO(task))
}

func (h *handler) handleList(w http.ResponseWriter) {
	tasks := h.repo.List()
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].UpdatedAt.Equal(tasks[j].UpdatedAt) {
			return tasks[i].ID < tasks[j].ID
		}
		return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
	})
	dtos := make([]TaskDTO, len(tasks))
	for i, t := range tasks {
		dtos[i] = toDTO(t)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dtos)
}

func (h *handler) handlePatch(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		Done *bool
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Done == nil {
		http.Error(w, "missing done field", http.StatusBadRequest)
		return
	}

	task, err := h.repo.SetDone(id, *req.Done)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toDTO(task))
}
