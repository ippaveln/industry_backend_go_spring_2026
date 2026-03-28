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

var errTaskNotFound = errors.New("task not found")

type InMemoryTaskRepo struct {
	mu     sync.RWMutex
	clock  Clock
	items  map[string]Task
	nextID uint64
}

func NewInMemoryTaskRepo(clock Clock) *InMemoryTaskRepo {
	return &InMemoryTaskRepo{
		clock: clock,
		items: make(map[string]Task),
	}
}

func (r *InMemoryTaskRepo) Create(title string) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, errors.New("title is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.nextID++
	id := strconv.FormatUint(r.nextID, 10)
	t := Task{
		ID:        id,
		Title:     title,
		Done:      false,
		UpdatedAt: r.clock.Now(),
	}
	r.items[id] = t
	return t, nil
}

func (r *InMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.items[id]
	return t, ok
}

func (r *InMemoryTaskRepo) List() []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Task, 0, len(r.items))
	for _, t := range r.items {
		out = append(out, t)
	}

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

	t, ok := r.items[id]
	if !ok {
		return Task{}, errTaskNotFound
	}

	t.Done = done
	t.UpdatedAt = r.clock.Now()
	r.items[id] = t

	return t, nil
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
			return
		case http.MethodGet:
			h.handleList(w, r)
			return
		default:
			http.NotFound(w, r)
			return
		}
	}

	id, ok := taskIDFromPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.handleGetByID(w, r, id)
	case http.MethodPatch:
		h.handlePatchDone(w, r, id)
	default:
		http.NotFound(w, r)
	}
}

func taskIDFromPath(path string) (string, bool) {
	const prefix = "/tasks/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	id := strings.TrimPrefix(path, prefix)
	if id == "" || strings.Contains(id, "/") {
		return "", false
	}
	return id, true
}

func (h *HTTPHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}
	if err := decodeStrictJSON(r, &req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	t, err := h.repo.Create(req.Title)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusCreated, toTaskDTO(t))
}

func (h *HTTPHandler) handleList(w http.ResponseWriter, _ *http.Request) {
	tasks := h.repo.List()
	out := make([]apiTaskDTO, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, toTaskDTO(t))
	}

	writeJSON(w, http.StatusOK, out)
}

func (h *HTTPHandler) handleGetByID(w http.ResponseWriter, r *http.Request, id string) {
	t, ok := h.repo.Get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}

	writeJSON(w, http.StatusOK, toTaskDTO(t))
}

func (h *HTTPHandler) handlePatchDone(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		Done *bool `json:"done"`
	}
	if err := decodeStrictJSON(r, &req); err != nil || req.Done == nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	t, err := h.repo.SetDone(id, *req.Done)
	if err != nil {
		if errors.Is(err, errTaskNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, toTaskDTO(t))
}

type apiTaskDTO struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func toTaskDTO(t Task) apiTaskDTO {
	return apiTaskDTO{
		ID:        t.ID,
		Title:     t.Title,
		Done:      t.Done,
		UpdatedAt: t.UpdatedAt,
	}
}

func decodeStrictJSON(r *http.Request, dst any) error {
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}

	var extra any
	if err := dec.Decode(&extra); err == nil {
		return errors.New("multiple JSON values")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
