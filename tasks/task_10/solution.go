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
	clock Clock
	seq   uint64
	tasks map[string]Task
}

func NewInMemoryTaskRepo(clock Clock) *InMemoryTaskRepo {
	return &InMemoryTaskRepo{
		clock: clock,
		tasks: make(map[string]Task),
	}
}

func (r *InMemoryTaskRepo) Create(title string) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, errors.New("empty title")
	}

	id := strconv.FormatUint(atomic.AddUint64(&r.seq, 1), 10)
	now := r.clock.Now()

	task := Task{
		ID:        id,
		Title:     title,
		Done:      false,
		UpdatedAt: now,
	}

	r.mu.Lock()
	r.tasks[id] = task
	r.mu.Unlock()

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
	out := make([]Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		out = append(out, task)
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

	task, ok := r.tasks[id]
	if !ok {
		return Task{}, errors.New("not found")
	}

	task.Done = done
	task.UpdatedAt = r.clock.Now()
	r.tasks[id] = task

	return task, nil
}

func NewHTTPHandler(repo TaskRepo) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/tasks", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			handleCreate(w, req, repo)
		case http.MethodGet:
			handleList(w, repo)
		default:
			http.NotFound(w, req)
		}
	})

	mux.HandleFunc("/tasks/", func(w http.ResponseWriter, req *http.Request) {
		id := strings.TrimPrefix(req.URL.Path, "/tasks/")
		if id == "" || strings.Contains(id, "/") {
			http.NotFound(w, req)
			return
		}

		switch req.Method {
		case http.MethodGet:
			handleGet(w, req, repo, id)
		case http.MethodPatch:
			handlePatch(w, req, repo, id)
		default:
			http.NotFound(w, req)
		}
	})

	return mux
}

type createTaskRequest struct {
	Title string `json:"title"`
}

type patchTaskRequest struct {
	Done *bool `json:"done"`
}

func handleCreate(w http.ResponseWriter, req *http.Request, repo TaskRepo) {
	var body createTaskRequest
	if !parseJSONBody(req, &body) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	task, err := repo.Create(body.Title)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

func handleGet(w http.ResponseWriter, req *http.Request, repo TaskRepo, id string) {
	task, ok := repo.Get(id)
	if !ok {
		http.NotFound(w, req)
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func handleList(w http.ResponseWriter, repo TaskRepo) {
	writeJSON(w, http.StatusOK, repo.List())
}

func handlePatch(w http.ResponseWriter, req *http.Request, repo TaskRepo, id string) {
	var body patchTaskRequest
	if !parseJSONBody(req, &body) || body.Done == nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	task, err := repo.SetDone(id, *body.Done)
	if err != nil {
		http.NotFound(w, req)
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func parseJSONBody(req *http.Request, dst any) bool {
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return false
	}

	if dec.More() {
		return false
	}

	var extra any
	if err := dec.Decode(&extra); err == nil {
		return false
	}

	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}