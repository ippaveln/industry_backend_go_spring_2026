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
	tasks     map[string]Task
	clock     Clock
	IDCounter int64
	mu        sync.RWMutex
}

func NewInMemoryTaskRepo(clock Clock) *InMemoryTaskRepo {
	return &InMemoryTaskRepo{
		tasks:     make(map[string]Task),
		clock:     clock,
		IDCounter: 0,
	}
}

func (r *InMemoryTaskRepo) nextID() string {
	r.IDCounter++
	return strconv.FormatInt(r.IDCounter, 10)
}

func (r *InMemoryTaskRepo) Create(title string) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, errors.New("title must not be empty")
	}

	id := r.nextID()
	timeNow := r.clock.Now()

	task := Task{
		ID:        id,
		Title:     title,
		Done:      false,
		UpdatedAt: timeNow,
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

func (r *InMemoryTaskRepo) List() []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]Task, 0, len(r.tasks))

	for _, task := range r.tasks {
		list = append(list, task)
	}

	sort.Slice(list, func(i, j int) bool {
		if !list[i].UpdatedAt.Equal(list[j].UpdatedAt) {
			return list[i].UpdatedAt.After(list[j].UpdatedAt)
		} else {
			return list[i].ID < list[j].ID
		}
	})

	return list
}

type httpHandler struct {
	repo TaskRepo
}

func NewHTTPHandler(repo TaskRepo) *httpHandler {
	return &httpHandler{repo: repo}
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/tasks" {
		switch r.Method {
		case http.MethodPost:
			h.createTask(w, r)
		case http.MethodGet:
			h.listTasks(w)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	if strings.HasPrefix(r.URL.Path, "/tasks/") {
		id := strings.TrimPrefix(r.URL.Path, "/tasks/")
		if id == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.getTask(w, id)
		case http.MethodPatch:
			h.setTaskDone(w, r, id)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	http.Error(w, "Not found", http.StatusNotFound)
}

func (h *httpHandler) createTask(w http.ResponseWriter, r *http.Request) {
	type createTaskIn struct {
		Title string `json:"title"`
	}
	var input createTaskIn

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&input)

	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		http.Error(w, "Title cannot be empty", http.StatusBadRequest)
		return
	}

	task, err := h.repo.Create(input.Title)
	if err != nil {
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *httpHandler) listTasks(w http.ResponseWriter) {
	tasks := h.repo.List()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

func (h *httpHandler) getTask(w http.ResponseWriter, id string) {
	task, ok := h.repo.Get(id)

	if !ok {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

func (h *httpHandler) setTaskDone(w http.ResponseWriter, r *http.Request, id string) {
	type doneStatusTask struct {
		Done *bool `json:"done"`
	}
	var input doneStatusTask

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&input)

	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if input.Done == nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	task, err := h.repo.SetDone(id, *input.Done)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}
