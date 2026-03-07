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

var (
	ErrUniqueID           error = errors.New("unique ID violation")
	ErrItemNotFound       error = errors.New("item not found")
	ErrInvalidRequestBody error = errors.New("invalid request body")
	ErrTitleRequired      error = errors.New("title required")
)

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateTaskReqParams struct {
	Title string `json:"title"`
}

type UpdateTaskReqParams struct {
	Done *bool `json:"done"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type TasksListResponse struct {
	Tasks []Task `json:"tasks"`
}

type Store struct {
	tasks  map[string]Task
	mu     sync.RWMutex
	clock  Clock
	nextID int
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

func NewInMemoryTaskRepo(clock Clock) *Store {
	return &Store{
		tasks: make(map[string]Task),
		clock: clock,
	}
}

func (s *Store) Create(title string) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	newID := strconv.Itoa(s.nextID)

	s.tasks[newID] = Task{
		ID:        newID,
		Title:     title,
		UpdatedAt: s.clock.Now(),
	}

	return s.tasks[newID], nil
}

func (s *Store) Get(id string) (Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if task, ok := s.tasks[id]; ok {
		return task, ok
	}
	return Task{}, false
}

func (s *Store) List() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]Task, 0, len(s.tasks))

	for _, v := range s.tasks {
		tasks = append(tasks, v)
	}

	sort.SliceStable(tasks, func(i, j int) bool {
		if tasks[i].UpdatedAt.Equal(tasks[j].UpdatedAt) {
			return tasks[i].ID < tasks[j].ID
		}
		return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
	})

	return tasks
}

func (s *Store) SetDone(id string, done bool) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, ok := s.tasks[id]; ok {
		task.Done = done
		task.UpdatedAt = s.clock.Now()
		s.tasks[id] = task
		return task, nil
	}

	return Task{}, ErrItemNotFound
}

func NewHTTPHandler(repo TaskRepo) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /tasks", func(w http.ResponseWriter, r *http.Request) {
		createTaskHandler(w, r, repo)
	})

	mux.HandleFunc("GET /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		getTaskHandler(w, r, repo)
	})

	mux.HandleFunc("PATCH /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		patchTaskHandler(w, r, repo)
	})

	mux.HandleFunc("GET /tasks", func(w http.ResponseWriter, r *http.Request) {
		getAllTasksHandler(w, r, repo)
	})

	return mux
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	if payload == nil {
		w.WriteHeader(code)
		return
	}

	data, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if _, err := w.Write(data); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func respondWithError(w http.ResponseWriter, code int, err error) {
	respondWithJSON(w, code, ErrorResponse{Error: err.Error()})
}

func createTaskHandler(w http.ResponseWriter, r *http.Request, repo TaskRepo) {
	reqCreateTask := CreateTaskReqParams{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&reqCreateTask); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrInvalidRequestBody)
		return
	}

	if strings.TrimSpace(reqCreateTask.Title) == "" {
		respondWithError(w, http.StatusBadRequest, ErrTitleRequired)
		return
	}

	createdTask, err := repo.Create(reqCreateTask.Title)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, createdTask)
}

func getTaskHandler(w http.ResponseWriter, r *http.Request, repo TaskRepo) {
	id := r.PathValue("id")
	if strings.TrimSpace(id) == "" {
		respondWithError(w, http.StatusBadRequest, ErrInvalidRequestBody)
		return
	}

	task, ok := repo.Get(id)
	if !ok {
		respondWithError(w, http.StatusNotFound, ErrItemNotFound)
		return
	}

	respondWithJSON(w, http.StatusOK, task)
}

func patchTaskHandler(w http.ResponseWriter, r *http.Request, repo TaskRepo) {
	id := r.PathValue("id")
	if strings.TrimSpace(id) == "" {
		respondWithError(w, http.StatusBadRequest, ErrInvalidRequestBody)
		return
	}

	reqUpdateTask := UpdateTaskReqParams{}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&reqUpdateTask); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrInvalidRequestBody)
		return
	}

	if reqUpdateTask.Done == nil {
		respondWithError(w, http.StatusBadRequest, ErrInvalidRequestBody)
		return
	}

	task, err := repo.SetDone(id, *reqUpdateTask.Done)
	if err != nil {
		respondWithError(w, http.StatusNotFound, ErrItemNotFound)
		return
	}

	respondWithJSON(w, http.StatusOK, task)
}

func getAllTasksHandler(w http.ResponseWriter, r *http.Request, repo TaskRepo) {
	tasks := repo.List()

	respondWithJSON(w, http.StatusOK, tasks)
}
