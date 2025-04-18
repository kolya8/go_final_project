package api

import (
	"net/http"

	"github.com/kolya8/go-sprint-thirteen/pkg/config"
)

var appConfig *config.Config

func RegisterHandlers(mux *http.ServeMux, cfg *config.Config) {
	appConfig = cfg

	mux.Handle("/", http.FileServer(http.Dir("web")))
	mux.HandleFunc("/api/nextdate", nextDateHandler)
	mux.HandleFunc("/api/task", authMiddleware(taskHandler))
	mux.HandleFunc("/api/tasks", authMiddleware(tasksHandler))
	mux.HandleFunc("/api/task/done", authMiddleware(taskDoneHandler))
	mux.HandleFunc("/api/signin", signinHandler)
}