package main

import (
	"encoding/json"
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

// REPO LVL

type inMemoryTaskRepo struct {
	tasks  map[string]Task
	clock  Clock
	nextID int
	mu     sync.RWMutex
}

func NewInMemoryTaskRepo(clock Clock) *inMemoryTaskRepo {
	return &inMemoryTaskRepo{
		tasks: make(map[string]Task),
		clock: clock,
	}
}

func (m *inMemoryTaskRepo) Create(title string) (Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nextID++

	task := Task{
		ID:        strconv.Itoa(m.nextID),
		Title:     title,
		Done:      false,
		UpdatedAt: m.clock.Now(),
	}

	m.tasks[task.ID] = task
	return task, nil
}

func (m *inMemoryTaskRepo) Get(id string) (Task, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, ok := m.tasks[id]
	return task, ok
}

func (m *inMemoryTaskRepo) List() []Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]Task, 0, len(m.tasks))

	for _, v := range m.tasks {
		tasks = append(tasks, v)
	}

	slices.SortFunc(tasks, func(a, b Task) int {
		if n := b.UpdatedAt.Compare(a.UpdatedAt); n != 0 {
			return n
		}
		// Сравнение ID как чисел, чтобы "10" > "2"
		idA, _ := strconv.Atoi(a.ID)
		idB, _ := strconv.Atoi(b.ID)
		return idA - idB
	})

	return tasks
}

func (m *inMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[id]
	if !ok {
		return Task{}, fmt.Errorf("task not found")
	}

	task.Done = done
	task.UpdatedAt = m.clock.Now()
	m.tasks[task.ID] = task

	return task, nil
}

// HANDLERS

type taskHandler struct {
	repo TaskRepo
}

func NewHTTPHandler(repo TaskRepo) http.Handler {
	mux := http.NewServeMux()
	h := &taskHandler{repo: repo}

	mux.HandleFunc("POST /tasks", h.create)
	mux.HandleFunc("GET /tasks", h.list)
	mux.HandleFunc("GET /tasks/{id}", h.get)
	mux.HandleFunc("PATCH /tasks/{id}", h.patch)

	return mux
}

func writeJSON[T any](w http.ResponseWriter, status int, task T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(task)
}

func (h *taskHandler) create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title string `json:"title"`
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(body.Title)
	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	task, err := h.repo.Create(title)
	if err != nil {
		http.Error(w, "error creating task", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

func (h *taskHandler) list(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.repo.List())
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

func (h *taskHandler) patch(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Done *bool `json:"done"`
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil || body.Done == nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	task, err := h.repo.SetDone(id, *body.Done)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, task)
}
