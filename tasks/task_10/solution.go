package main

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateRandomID() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

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

type taskHandler struct {
	taskRepo TaskRepo
}

type taskRepo struct {
	clock   Clock
	storage map[string]Task
	mu      sync.RWMutex
}

func (tr *taskRepo) Create(title string) (Task, error) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	task := Task{
		ID:        generateRandomID(),
		Title:     title,
		Done:      false,
		UpdatedAt: tr.clock.Now(),
	}
	tr.storage[task.ID] = task
	return task, nil
}

func (tr *taskRepo) Get(id string) (Task, bool) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	task, exists := tr.storage[id]
	return task, exists
}

func (tr *taskRepo) List() []Task {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	taskList := make([]Task, 0, len(tr.storage))
	for _, task := range tr.storage {
		taskList = append(taskList, task)
	}

	sort.Slice(taskList, func(i, j int) bool {
		return taskList[i].UpdatedAt.Unix() > taskList[j].UpdatedAt.Unix() || (taskList[i].UpdatedAt.Unix() == taskList[j].UpdatedAt.Unix() && taskList[i].ID < taskList[j].ID)
	})
	return taskList
}

func (tr *taskRepo) SetDone(id string, done bool) (Task, error) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	if task, exists := tr.storage[id]; exists {
		newTask := Task{
			ID:        task.ID,
			Title:     task.Title,
			Done:      done,
			UpdatedAt: tr.clock.Now(),
		}
		tr.storage[task.ID] = newTask
		return newTask, nil
	} else {
		return task, errors.New("entity not found")
	}
}

func NewHTTPHandler(taskRepo TaskRepo) *taskHandler {
	return &taskHandler{taskRepo: taskRepo}
}

func NewInMemoryTaskRepo(clock Clock) *taskRepo {
	return &taskRepo{clock: clock, storage: make(map[string]Task)}
}

func (t *taskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == "/tasks" {
		t.GetTasks(w, r)
		return
	}
	if r.Method == http.MethodPost && r.URL.Path == "/tasks" {
		t.CreateTask(w, r)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/tasks/")

	if r.Method == http.MethodGet {
		t.GetTask(w, r, id)
		return
	}
	if r.Method == http.MethodPatch {
		t.UpdateDone(w, r, id)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (t *taskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	tasks := t.taskRepo.List()
	jsonBytes, err := json.Marshal(tasks)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (t *taskHandler) GetTask(w http.ResponseWriter, r *http.Request, id string) {
	task, exists := t.taskRepo.Get(id)
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	jsonBytes, err := json.Marshal(task)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (t *taskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	trimmedTitle := strings.TrimSpace(req.Title)
	if trimmedTitle == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	task, err := t.taskRepo.Create(trimmedTitle)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if jsonBytes, err := json.Marshal(task); err != nil {
		w.WriteHeader(http.StatusBadRequest)

	} else {
		w.WriteHeader(http.StatusCreated)
		w.Write(jsonBytes)
	}

}

type createTaskRequest struct {
	Title string `json:"title"`
}

type setDoneRequest struct {
	Done *bool `json:"done"`
}

func (t *taskHandler) UpdateDone(w http.ResponseWriter, r *http.Request, id string) {
	var req setDoneRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.Done == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	task, err := t.taskRepo.SetDone(id, *req.Done)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	jsonBytes, err := json.Marshal(task)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}
