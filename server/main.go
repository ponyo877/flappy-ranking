package main

import (
	"log"
	"net/http"

	"github.com/ponyo877/flappy-standings/server/adapter"
	"github.com/ponyo877/flappy-standings/server/database"
	"github.com/ponyo877/flappy-standings/server/repository"
	"github.com/ponyo877/flappy-standings/server/usecase"
)

func main() {
	db, err := database.NewMySQL()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	repository := repository.NewScoreRepository(db)
	usecase := usecase.NewScoreUsecase(repository)
	adapter := adapter.NewAdapter(usecase)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /tokens", adapter.GenerateTokenHandler)
	mux.HandleFunc("GET /scores", adapter.ListScoreHandler)
	mux.HandleFunc("POST /scores", adapter.RegisterScoreHandler)
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
