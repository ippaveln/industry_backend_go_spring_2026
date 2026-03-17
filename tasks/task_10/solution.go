package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

var errNotFound = errors.New("task not found")

type InMemoryTaskRepo struct {
	mu    sync.RWMutex
	data  map[string]Task
	clock Clock
}

var idSeq uint64

func NewInMemoryTaskRepo(clock Clock) *InMemoryTaskRepo {
	return &InMemoryTaskRepo{
		data:  make(map[string]Task),
		clock: clock,
	}
}

func nextID() string {
	id := atomic.AddUint64(&idSeq, 1)
	return fmt.Sprintf("task-%d", id)
}

func (r *InMemoryTaskRepo) Create(title string) (Task, error) {
	id := nextID()
	t := Task{ID: id, Title: title, Done: false, UpdatedAt: r.clock.Now()}

	r.mu.Lock()
	r.data[id] = t
	r.mu.Unlock()

	return t, nil
}

func (r *InMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mu.RLock()
	t, ok := r.data[id]
	r.mu.RUnlock()
	return t, ok
}

func (r *InMemoryTaskRepo) List() []Task {
	r.mu.RLock()
	out := make([]Task, 0, len(r.data))
	for _, t := range r.data {
		out = append(out, t)
	}
	r.mu.RUnlock()

	sort.Slice(out, func(i, j int) bool {
		if out[i].UpdatedAt.Equal(out[j].UpdatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})

	return out
}

func (r *InMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.data[id]
	if !ok {
		return Task{}, errNotFound
	}

	t.Done = done
	t.UpdatedAt = r.clock.Now()
	r.data[id] = t

	return t, nil
}

type HTTPHandler struct {
	repo TaskRepo
	mux  *http.ServeMux
}

func NewHTTPHandler(repo TaskRepo) http.Handler {
	h := &HTTPHandler{repo: repo, mux: http.NewServeMux()}
	h.mux.HandleFunc("/tasks", h.handleTasks)
	h.mux.HandleFunc("/tasks/", h.handleTaskByID)
	return h
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *HTTPHandler) handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createTask(w, r)
	case http.MethodGet:
		h.listTasks(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *HTTPHandler) handleTaskByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if id == "" || strings.Contains(id, "/") {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getTask(w, r, id)
	case http.MethodPatch:
		h.patchTask(w, r, id)
	default:
		http.NotFound(w, r)
	}
}

func (h *HTTPHandler) createTask(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Title string `json:"title"`
	}

	var req reqBody
	if !decodeStrictJSON(r, &req) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	t, err := h.repo.Create(title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, t)
}

func (h *HTTPHandler) getTask(w http.ResponseWriter, _ *http.Request, id string) {
	t, ok := h.repo.Get(id)
	if !ok {
		http.NotFound(w, nil)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *HTTPHandler) listTasks(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, h.repo.List())
}

func (h *HTTPHandler) patchTask(w http.ResponseWriter, r *http.Request, id string) {
	type reqBody struct {
		Done *bool `json:"done"`
	}

	var req reqBody
	if !decodeStrictJSON(r, &req) || req.Done == nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	t, err := h.repo.SetDone(id, *req.Done)
	if err != nil {
		if errors.Is(err, errNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, t)
}

func decodeStrictJSON(r *http.Request, dst any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return false
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
