package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
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

type InMemoryTaskRepo struct {
	clock Clock
	mu    sync.RWMutex
	data  map[string]Task
}

func decodeStrictJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

type patchReq struct {
	Done *bool `json:"done"`
}

type TaskRepo interface {
	Create(title string) (Task, error)
	Get(id string) (Task, bool)
	List() []Task
	SetDone(id string, done bool) (Task, error)
}

type HTTPHandler struct {
	repo TaskRepo
	mux  *http.ServeMux
}

func NewHTTPHandler(repo TaskRepo) *HTTPHandler {
	h := &HTTPHandler{
		repo: repo,
		mux:  http.NewServeMux(),
	}

	h.mux.HandleFunc("/tasks", h.tasks)
	h.mux.HandleFunc("/tasks/", h.tasks)

	return h
}

func (h *HTTPHandler) tasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getTasks(w, r)
	case http.MethodPost:
		h.createTask(w, r)
	case http.MethodPatch:
		h.patchTask(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *HTTPHandler) getTasks(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/tasks")

	if rest == "" || rest == "/" {
		tasks := h.repo.List()
		writeJSON(w, http.StatusOK, tasks)
		return
	}

	if strings.HasPrefix(rest, "/") {
		rest = rest[1:]
	}

	task, ok := h.repo.Get(rest)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *HTTPHandler) createTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	if err := decodeStrictJSON(r, &task); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	task, err := h.repo.Create(task.Title)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (h *HTTPHandler) patchTask(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	var req patchReq
	if err := decodeStrictJSON(r, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.Done == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	task, err := h.repo.SetDone(id, *req.Done)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

var nextID uint64

func NewID() string {
	id := atomic.AddUint64(&nextID, 1)
	return strconv.FormatUint(id, 10)
}

func NewInMemoryTaskRepo(clock Clock) *InMemoryTaskRepo {
	return &InMemoryTaskRepo{
		clock: clock,
		data:  make(map[string]Task),
	}
}

func (r *InMemoryTaskRepo) Create(title string) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, errors.New("title is empty")
	}
	id := NewID()
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
	if !ok {
		return Task{}, false
	}
	return task, true
}

func (r *InMemoryTaskRepo) List() []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]Task, 0, len(r.data))

	for _, task := range r.data {
		list = append(list, task)
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].UpdatedAt.Equal(list[j].UpdatedAt) {
			return list[i].ID < list[j].ID
		}
		return list[i].UpdatedAt.After(list[j].UpdatedAt)
	})
	listCopy := make([]Task, len(list))
	copy(listCopy, list)

	return listCopy
}
func (r *InMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.data[id]
	if !ok {
		return Task{}, errors.New("task not found")
	}

	task.Done = done
	task.UpdatedAt = r.clock.Now()

	r.data[id] = task

	return task, nil
}
