package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"fmt"

	"github.com/kolya8/go-sprint-thirteen/pkg/db"
)

type SuccessResponse struct {
	ID string `json:"id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		sendResponse(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "unsupported request method"})
	}
}

func taskDoneHandler(w http.ResponseWriter, r *http.Request) {
	task, err := getTask(r)
	if err != nil {
		sendResponse(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	var responseErr error

	if task.Repeat == "" {
		responseErr = db.DeleteTask(task.ID)
	} else {
		now := time.Now()
		nextDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			responseErr = err
		} else {
			responseErr = db.UpdateDate(nextDate, task.ID)
		}
	}

	if responseErr != nil {
		sendResponse(w, http.StatusBadRequest, ErrorResponse{Error: responseErr.Error()})
		return
	}

	sendResponse(w, http.StatusOK, struct{}{})
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		sendResponse(w, http.StatusBadRequest, ErrorResponse{Error: "id is required"})
		return
	}

	err := db.DeleteTask(id)
	if err != nil {
		sendResponse(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	sendResponse(w, http.StatusOK, struct{}{})
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	task, err := taskValidate(r)
	if err != nil {
		sendResponse(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	err = db.UpdateTask(task)
	if err != nil {
		sendResponse(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	sendResponse(w, http.StatusOK, struct{}{})
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	task, err := getTask(r)
	if err != nil {
		sendResponse(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	sendResponse(w, http.StatusOK, task)
}

func getTask(r *http.Request) (*db.Task, error) {
	id := r.URL.Query().Get("id")
	if id == "" {
		return nil, errors.New("id is required")
	}

	task, err := db.GetTask(id)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {

	task, err := taskValidate(r)
	if err != nil {
		sendResponse(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	id, err := db.AddTask(task)
	if err != nil {
		sendResponse(w, http.StatusBadRequest, ErrorResponse{Error: "db: add task error"})
		return
	}
	sendResponse(w, http.StatusOK, SuccessResponse{ID: fmt.Sprintf("%d", id)})
}

func taskValidate(r *http.Request) (*db.Task, error) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.New("request body reading error")
	}
	
	var task db.Task

	err = json.Unmarshal(body, &task)
	if err != nil {
		return nil, errors.New("JSON parsing error")
	}

	if task.Title == "" {
		return nil, errors.New("empty Title")
	}

	err = checkDate(&task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func checkDate(task *db.Task) error {
	now := time.Now()
	today := now.Format(dateFormat)

	if task.Date == "" {
		task.Date = today
	}

	date, err := time.Parse(dateFormat, task.Date)
	if err != nil {
		return err
	}

	startToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if afterNow(startToday, date) {
		if len(task.Repeat) != 0 {
			nextDate, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return err
			}
			task.Date = nextDate
		} else {
			task.Date = today
		}
	}

	return nil
}

func sendResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
