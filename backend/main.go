package main

import (
	"log"
	"net/http"

	"github.com/ponyo877/flappy-standings/backend/database"
	"github.com/ponyo877/flappy-standings/backend/repository"
	"github.com/ponyo877/flappy-standings/backend/server"
	"github.com/ponyo877/flappy-standings/backend/usecase"
)

func main() {
	db, err := database.NewMySQL()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	repository := repository.NewScoreRepository(db)
	usecase := usecase.NewScoreUsecase(repository)
	server := server.NewServer(usecase)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /tokens", server.GenerateTokenHandler)
	mux.HandleFunc("GET /scores", server.ListScoreHandler)
	mux.HandleFunc("POST /scores", server.RegisterScoreHandler)
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
