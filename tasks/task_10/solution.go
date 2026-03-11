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
type Task struct {
	ID        string
	Title     string
	Done      bool
	UpdatedAt time.Time
}

type Clock interface {
	Now() time.Time
}

type TaskRepo interface {
	Create(title string) (Task, error)
	Get(id string) (Task, bool)
	List() []Task
	SetDone(id string, done bool) (Task, error)
}

type inMemoryTaskRepo struct {
	mu      sync.RWMutex
	tasks   map[string]Task
	clock   Clock
	counter int64
}

func NewInMemoryTaskRepo(clock Clock) *inMemoryTaskRepo {
	return &inMemoryTaskRepo{
		tasks:   make(map[string]Task),
		clock:   clock,
		counter: 0,
	}
}

func (r *inMemoryTaskRepo) Create(title string) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, errors.New("title cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.counter++
	id := strconv.FormatInt(r.counter, 10) 

	task := Task{
		ID:        id,
		Title:     title,
		Done:      false,
		UpdatedAt: r.clock.Now(),
	}

	r.tasks[id] = task
	return task, nil
}

func (r *inMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]
	return task, ok
}

func (r *inMemoryTaskRepo) List() []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}

	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].UpdatedAt.Equal(tasks[j].UpdatedAt) {
			return tasks[i].ID < tasks[j].ID
		}
		return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
	})

	return tasks
}

func (r *inMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
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

type TaskHandler struct {
	repo TaskRepo
}

func NewHTTPHandler(repo TaskRepo) http.Handler {
	mux := http.NewServeMux()
	h := &TaskHandler{repo: repo}

	mux.HandleFunc("POST /tasks", h.createTask)
	mux.HandleFunc("GET /tasks/{id}", h.getTask)
	mux.HandleFunc("GET /tasks", h.listTasks)
	mux.HandleFunc("PATCH /tasks/{id}", h.updateTask)

	return mux
}

func (h *TaskHandler) writeTask(w http.ResponseWriter, status int, task Task) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(struct {
		ID        string    `json:"id"`
		Title     string    `json:"title"`
		Done      bool      `json:"done"`
		UpdatedAt time.Time `json:"updatedAt"`
	}{
		ID:        task.ID,
		Title:     task.Title,
		Done:      task.Done,
		UpdatedAt: task.UpdatedAt,
	})
}

func (h *TaskHandler) createTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	task, err := h.repo.Create(req.Title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.writeTask(w, http.StatusCreated, task)
}

func (h *TaskHandler) getTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	task, ok := h.repo.Get(id)
	if !ok {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	h.writeTask(w, http.StatusOK, task)
}

func (h *TaskHandler) listTasks(w http.ResponseWriter, r *http.Request) {
	tasks := h.repo.List()

	dtoList := make([]interface{}, len(tasks))
	for i, task := range tasks {
		dtoList[i] = struct {
			ID        string    `json:"id"`
			Title     string    `json:"title"`
			Done      bool      `json:"done"`
			UpdatedAt time.Time `json:"updatedAt"`
		}{
			ID:        task.ID,
			Title:     task.Title,
			Done:      task.Done,
			UpdatedAt: task.UpdatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dtoList)
}

func (h *TaskHandler) updateTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req struct {
		Done *bool `json:"done"` 
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Done == nil {
		http.Error(w, "Missing required field: done", http.StatusBadRequest)
		return
	}

	task, err := h.repo.SetDone(id, *req.Done)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	h.writeTask(w, http.StatusOK, task)
}
