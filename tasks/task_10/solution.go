package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	BadRequest error = errors.New("BadRequest")
	NotFound   error = errors.New("Not Found")
)

type Clock interface {
	Now() time.Time
}

type Repo struct {
	data  map[string]Task
	clock Clock
	mutex *sync.RWMutex
}

type TaskRepo interface {
	Create(title string) (Task, error)
	Get(id string) (Task, bool)
	List() []Task
	SetDone(id string, done bool) (Task, error)
}

type Task struct {
	ID        string
	Title     string
	Done      bool
	UpdatedAt time.Time
}

type TaskHandler struct{}

func NewInMemoryTaskRepo(clock Clock) *Repo {
	data := make(map[string]Task)
	return &Repo{
		data:  data,
		clock: clock,
		mutex: &sync.RWMutex{},
	}
}

func GenerateID(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (r *Repo) Create(title string) (Task, error) {
	title = strings.TrimSpace(title)
	if len(title) == 0 {
		return Task{}, BadRequest
	}
	id, _ := GenerateID(16)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	task := Task{
		ID:        id,
		Title:     title,
		Done:      false,
		UpdatedAt: r.clock.Now(),
	}
	r.data[id] = task

	return task, nil
}

func (r *Repo) Get(id string) (Task, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	task, exists := r.data[id]
	return task, exists
}

func (r *Repo) List() []Task {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tasks := make([]Task, 0, len(r.data))
	for _, task := range r.data {
		tasks = append(tasks, task)
	}
	sort.Slice(tasks, func(i, j int) bool {
		if !tasks[i].UpdatedAt.Equal(tasks[j].UpdatedAt) {
			return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
		}
		return tasks[i].ID < tasks[j].ID
	})
	return tasks
}

func (r *Repo) SetDone(id string, done bool) (Task, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	task, exists := r.data[id]
	if !exists {
		return Task{}, NotFound
	}
	task.Done = done
	task.UpdatedAt = r.clock.Now()
	r.data[id] = task

	return task, nil
}

type taskResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func toResponse(t Task) taskResponse {
	return taskResponse{ID: t.ID, Title: t.Title, Done: t.Done, UpdatedAt: t.UpdatedAt}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func NewHTTPHandler(repo *Repo) http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc("POST /tasks", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Title string `json:"title"`
		}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		task, err := repo.Create(req.Title)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusCreated, toResponse(task))
	})

	mux.HandleFunc("GET /tasks", func(w http.ResponseWriter, r *http.Request) {
		tasks := repo.List()
		resp := make([]taskResponse, len(tasks))
		for i, t := range tasks {
			resp[i] = toResponse(t)
		}
		writeJSON(w, http.StatusOK, resp)
	})

	mux.HandleFunc("GET /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		task, ok := repo.Get(id)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, toResponse(task))
	})

	mux.HandleFunc("PATCH /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var req struct {
			Done *bool `json:"done"`
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
		task, err := repo.SetDone(id, *req.Done)
		if err != nil {
			if errors.Is(err, NotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, toResponse(task))
	})

	return mux
}
