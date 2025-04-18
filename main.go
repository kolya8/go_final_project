package main

import (
	"log"

	"github.com/kolya8/go-sprint-thirteen/pkg/server"
)

func main() {
	err := server.Run()
	if err != nil {
		log.Fatalf("server error: %v", err)
	}
}
