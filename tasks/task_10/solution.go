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
	mu    sync.RWMutex
	tasks map[string]Task
	seq   uint64
	clock Clock
}

func NewInMemoryTaskRepo(clock Clock) *InMemoryTaskRepo {
	return &InMemoryTaskRepo{
		tasks: make(map[string]Task),
		clock: clock,
	}
}

func (r *InMemoryTaskRepo) nextID() string {
	r.seq++
	return strconv.FormatUint(r.seq, 10)
}

func (r *InMemoryTaskRepo) Create(title string) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := r.nextID()
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

	res := make([]Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		res = append(res, t)
	}

	sort.Slice(res, func(i, j int) bool {
		if !res[i].UpdatedAt.Equal(res[j].UpdatedAt) {
			return res[i].UpdatedAt.After(res[j].UpdatedAt)
		}
		return res[i].ID < res[j].ID
	})

	return res
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

type HTTPHandler struct {
	repo TaskRepo
}

func NewHTTPHandler(repo TaskRepo) http.Handler {
	return &HTTPHandler{repo: repo}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path == "/tasks" {
		switch r.Method {
		case http.MethodPost:
			h.handleCreate(w, r)
		case http.MethodGet:
			h.handleList(w, r)
		default:
			http.NotFound(w, r)
		}
		return
	}

	if strings.HasPrefix(r.URL.Path, "/tasks/") {
		id := strings.TrimPrefix(r.URL.Path, "/tasks/")
		switch r.Method {
		case http.MethodGet:
			h.handleGet(w, r, id)
		case http.MethodPatch:
			h.handlePatch(w, r, id)
		default:
			http.NotFound(w, r)
		}
		return
	}

	http.NotFound(w, r)
}

type createReq struct {
	Title string `json:"title"`
}

func (h *HTTPHandler) handleCreate(w http.ResponseWriter, r *http.Request) {

	var req createReq

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	task, err := h.repo.Create(title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

func (h *HTTPHandler) handleGet(w http.ResponseWriter, r *http.Request, id string) {

	task, ok := h.repo.Get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (h *HTTPHandler) handleList(w http.ResponseWriter, r *http.Request) {

	list := h.repo.List()
	writeJSON(w, http.StatusOK, list)
}

type patchReq struct {
	Done *bool `json:"done"`
}

func (h *HTTPHandler) handlePatch(w http.ResponseWriter, r *http.Request, id string) {

	var req patchReq

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.Done == nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	task, err := h.repo.SetDone(id, *req.Done)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
