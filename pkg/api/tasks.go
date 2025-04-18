package api

import (
	"net/http"
	"time"

	"github.com/kolya8/go-sprint-thirteen/pkg/db"
)

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendResponse(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "unsupported request method"})
		return
	}

	tasksQuantity := 50
	search := r.URL.Query().Get("search")

	var tasks *db.TasksResp
	var err error

	if search != "" {
		formattedDate, dateErr := isValidDate(search)
		if dateErr == nil { 
			tasks, err = db.GetTasksByDate(tasksQuantity, formattedDate)
		} else {
			tasks, err = db.GetTasksBySearchWord(tasksQuantity, search)
		}
	} else {
		tasks, err = db.Tasks(tasksQuantity)
	}

	if err != nil {
		sendResponse(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	sendResponse(w, http.StatusOK, tasks)
}

func isValidDate(dateStr string) (string, error) {
	const layout = "02.01.2006"
	date, err := time.Parse(layout, dateStr)
	if err != nil {
		return "", err
	}
	formattedDate := date.Format(dateFormat)
	return formattedDate, nil
}
