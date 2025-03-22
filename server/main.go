//go:build js && wasm

package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/ponyo877/flappy-ranking/server/adapter"
	"github.com/ponyo877/flappy-ranking/server/repository"
	"github.com/ponyo877/flappy-ranking/server/usecase"
	"github.com/syumai/workers"

	_ "github.com/syumai/workers/cloudflare/d1" // register driver
)

const dbName = "FlappyDB"

func main() {
	db, err := sql.Open("d1", dbName)
	if err != nil {
		log.Printf("failed to connect to database: %v", err)
		return
	}

	repository := repository.NewScoreRepository(db)
	usecase := usecase.NewScoreUsecase(repository)
	adapter := adapter.NewAdapter(usecase)

	http.HandleFunc("POST /api/tokens", adapter.GenerateTokenHandler)
	http.HandleFunc("GET /api/scores", adapter.ListScoreHandler)
	http.HandleFunc("POST /api/scores/{token}", adapter.RegisterScoreHandler)
	http.HandleFunc("POST /api/sessions/{token}", adapter.FinishSessionHandler)

	workers.Serve(nil)
}
