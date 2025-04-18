package server

import (
	"fmt"
	"net/http"

	"github.com/kolya8/go-sprint-thirteen/pkg/api"
	"github.com/kolya8/go-sprint-thirteen/pkg/config"
	"github.com/kolya8/go-sprint-thirteen/pkg/db"
)

func Run() error {
	config := config.Load()

	err := db.InitDB(config.DBFile)
	if err != nil {
		return fmt.Errorf("database initialization error: %w", err)
	}

	mux := http.NewServeMux()
	api.RegisterHandlers(mux, &config)

	fmt.Printf("starting server on port %s\n", config.Port)
	return http.ListenAndServe(":"+config.Port, mux)
}