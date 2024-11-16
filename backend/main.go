package main

import (
	"net/http"

	"github.com/ponyo877/flappy-standings/backend/repository"
	"github.com/ponyo877/flappy-standings/backend/server"
	"github.com/ponyo877/flappy-standings/backend/usecase"
)

func main() {
	repository := repository.NewScoreRepository(nil)
	usecase := usecase.NewScoreUsecase(repository)
	server := server.NewServer(usecase)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /token", server.TokenHandler)
	mux.HandleFunc("GET /score", server.ScoreHandler)
	mux.HandleFunc("POST /score", server.ScoreRegisterHandler)
	http.ListenAndServe(":8080", mux)
}
