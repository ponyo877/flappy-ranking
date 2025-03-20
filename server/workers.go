//go:build js && wasm

package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/ponyo877/flappy-standings/server/adapter"
	"github.com/ponyo877/flappy-standings/server/repository"
	"github.com/ponyo877/flappy-standings/server/usecase"
	"github.com/syumai/workers"

	_ "github.com/syumai/workers/cloudflare/d1" // register driver
)

const dbName = "flappy-db"

func main() {
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
	db, err := sql.Open("d1", dbName)
	if err != nil {
		log.Printf("failed to connect to database: %v", err)
		return
	}

	repository := repository.NewScoreRepository(db)
	usecase := usecase.NewScoreUsecase(repository)
	adapter := adapter.NewAdapter(usecase)

	http.HandleFunc("POST /tokens", adapter.GenerateTokenHandler)
	http.HandleFunc("GET /scores", adapter.ListScoreHandler)
	http.HandleFunc("POST /scores/{token}", adapter.RegisterScoreHandler)

	handler := corsMiddleware(http.DefaultServeMux)
	workers.Serve(handler)
}
