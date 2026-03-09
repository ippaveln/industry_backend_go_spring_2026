package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
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

type errorDTO struct {
	Error string `json:"error"`
}

// NewHTTPHandler FIXME: inject logger throw interface
func NewHTTPHandler(repo *TaskRepoImpl) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /tasks", func(w http.ResponseWriter, r *http.Request) {
		handleList(w, repo)
	})

	mux.HandleFunc("POST /tasks", func(writer http.ResponseWriter, request *http.Request) {
		handleCreate(writer, request, repo)
	})

	mux.HandleFunc("GET /tasks/{id}", func(writer http.ResponseWriter, request *http.Request) {
		handleGet(writer, request, repo)
	})

	mux.HandleFunc("PATCH /tasks/{id}", func(writer http.ResponseWriter, request *http.Request) {
		handleUpdate(writer, request, repo)
	})

	return mux
}

func handleUpdate(writer http.ResponseWriter, request *http.Request, repo *TaskRepoImpl) {
	id := request.PathValue("id")

	var input struct {
		IsDone *bool `json:"done"`
	}

	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&input)
	if err != nil || input.IsDone == nil {
		slog.Warn("wrong input json on path request", "error", err)
		writeError(writer, "wrong input json", http.StatusBadRequest)
		return
	}

	task, err := repo.SetDone(id, *input.IsDone)
	if err != nil {
		slog.Warn("task with this ID doesn't exist", "error", err)
		writeError(writer, "task with this ID doesn't exist", http.StatusNotFound)
		return
	}

	writer.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(writer).Encode(task)
	if err != nil {
		slog.Error("failed to encode response", "error", err)
		writeError(writer, "something went wrong", http.StatusInternalServerError)
		return
	}
	slog.Info("task updated", "id", id, "done", *input.IsDone)
}

func handleGet(writer http.ResponseWriter, request *http.Request, repo *TaskRepoImpl) {
	id := request.PathValue("id")

	task, isExists := repo.Get(id)

	if !isExists {
		slog.Warn("haven't found task with requested id on get request")
		writeError(writer, "task with this ID doesn't exist", http.StatusNotFound)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(writer).Encode(task)

	if err != nil {
		slog.Error("failed to encode response", "error", err)
		writeError(writer, "something went wrong", http.StatusInternalServerError)
		return
	}
}

func handleCreate(writer http.ResponseWriter, request *http.Request, repo *TaskRepoImpl) {
	var input struct {
		Title string `json:"title"`
	}

	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&input)
	if err != nil {
		slog.Warn("got bad input json on create task", "error", err)
		writeError(writer, "wrong input json", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(input.Title) == "" {
		slog.Warn("got empty input json on create task", "error", err)
		writeError(writer, "empty input json", http.StatusBadRequest)
		return
	}

	createdTask, err := repo.Create(input.Title)
	if err != nil {
		slog.Error("failed to generate UUID while creating task", "error", err)
		writeError(writer, "something went wrong", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(writer).Encode(createdTask)

	if err != nil {
		slog.Error("failed to encode response", "error", err)
		writeError(writer, "something went wrong", http.StatusInternalServerError)
		return
	}
	slog.Info("task created", "id", createdTask.ID)
}

func handleList(writer http.ResponseWriter, repo *TaskRepoImpl) {
	tasks := repo.List()

	sort.SliceStable(tasks, func(i, j int) bool {
		if tasks[i].UpdatedAt.Equal(tasks[j].UpdatedAt) {
			return tasks[i].ID < tasks[j].ID
		}
		return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
	})

	writer.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(writer).Encode(tasks)

	if err != nil {
		slog.Error("failed to encode response", "error", err)
		writeError(writer, "something went wrong", http.StatusInternalServerError)
		return
	}
}

func writeError(writer http.ResponseWriter, message string, status int) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)

	err := json.NewEncoder(writer).Encode(errorDTO{Error: message})
	if err != nil {
		slog.Error("failed to encode error response", "error", err)
		return
	}
}

func generateUUID() (string, error) {
	b := make([]byte, 16)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

type TaskRepoImpl struct {
	mutex     sync.RWMutex
	valuesMap map[string]*Task
	clock     Clock
}

func NewInMemoryTaskRepo(clock Clock) *TaskRepoImpl {
	return &TaskRepoImpl{
		clock:     clock,
		valuesMap: make(map[string]*Task),
	}
}

func (t *TaskRepoImpl) Create(title string) (Task, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	taskId, err := generateUUID()

	if err != nil {
		return Task{}, err
	}

	currTime := t.clock.Now()

	newTask := Task{
		ID:        taskId,
		Title:     title,
		Done:      false,
		UpdatedAt: currTime,
	}
	t.valuesMap[taskId] = &newTask

	return newTask, nil
}

func (t *TaskRepoImpl) Get(id string) (Task, bool) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	val, isExists := t.valuesMap[id]

	if !isExists {
		return Task{}, false
	}

	return *val, isExists
}

func (t *TaskRepoImpl) List() []Task {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	result := make([]Task, 0, len(t.valuesMap))

	for _, value := range t.valuesMap {
		elCopy := Task{
			value.ID,
			value.Title,
			value.Done,
			value.UpdatedAt,
		}

		result = append(result, elCopy)
	}

	return result
}

func (t *TaskRepoImpl) SetDone(id string, done bool) (Task, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	val, isExists := t.valuesMap[id]

	if !isExists {
		return Task{}, errors.New("task hasn't been found in repo")
	}

	currTime := t.clock.Now()

	val.Done = done
	val.UpdatedAt = currTime

	return *val, nil
}
