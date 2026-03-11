package main

import (
	"encoding/json"
	"errors"
	"fmt"
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
	store  map[string]Task
	nextID int
	mu     sync.RWMutex
	clock  Clock
}

func NewInMemoryTaskRepo(clock Clock) *InMemoryTaskRepo {
	return &InMemoryTaskRepo{store: make(map[string]Task), clock: clock}
}

func (r *InMemoryTaskRepo) Create(title string) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if strings.TrimSpace(title) == "" {
		return Task{}, errors.New("title cannot be empty")
	}

	taskID := strconv.Itoa(r.nextID)
	r.nextID++

	task := Task{ID: taskID, Title: title, Done: false, UpdatedAt: r.clock.Now()}
	r.store[taskID] = task

	return task, nil
}

func (r *InMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.store[id]

	return task, ok
}

func (r *InMemoryTaskRepo) List() []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]Task, 0, len(r.store))
	for _, v := range r.store {
		tasks = append(tasks, v)
	}

	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].UpdatedAt.Equal(tasks[j].UpdatedAt) {
			iIDString, _ := strconv.Atoi(tasks[i].ID)
			jIDString, _ := strconv.Atoi(tasks[j].ID)

			return iIDString < jIDString
		}

		return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
	})

	return tasks
}

func (r *InMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.store[id]
	if !ok {
		return Task{}, fmt.Errorf("task %q not found", id)
	}

	task.Done = done
	task.UpdatedAt = r.clock.Now()
	r.store[id] = task

	return task, nil
}

type HTTPHandler struct {
	repo TaskRepo
}

func NewHTTPHandler(repo TaskRepo) *HTTPHandler {
	return &HTTPHandler{repo: repo}
}

type CreateTask struct {
	Title string `json:"title"`
}

type UpdateTask struct {
	Done *bool `json:"done"`
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method

	switch {
	case path == "/tasks" && method == http.MethodPost:
		h.handleCreateTask(w, r)
	case path == "/tasks" && method == http.MethodGet:
		h.handleListTasks(w)
	case strings.HasPrefix(path, "/tasks/") && method == http.MethodGet:
		id := strings.TrimPrefix(path, "/tasks/")
		h.handleGetTask(w, id)
	case strings.HasPrefix(path, "/tasks/") && method == http.MethodPatch:
		id := strings.TrimPrefix(path, "/tasks/")
		h.handleUpdateTask(w, r, id)
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func (h *HTTPHandler) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var task CreateTask

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&task)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(task.Title) == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	createdTask, err := h.repo.Create(task.Title)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdTask)
}

func (h *HTTPHandler) handleListTasks(w http.ResponseWriter) {
	tasks := h.repo.List()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

func (h *HTTPHandler) handleGetTask(w http.ResponseWriter, id string) {
	task, ok := h.repo.Get(id)

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

func (h *HTTPHandler) handleUpdateTask(w http.ResponseWriter, r *http.Request, id string) {
	var task UpdateTask

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&task)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if task.Done == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updatedTask, err := h.repo.SetDone(id, *task.Done)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedTask)
}
