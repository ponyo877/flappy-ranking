package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ponyo877/flappy-standings/common"
)

func main() {
	server := NewServer(NewScoreUsecase())
	mux := http.NewServeMux()
	mux.HandleFunc("POST /score", server.scoreHandler)
	http.ListenAndServe(":8080", mux)
}

type Server struct {
	usecase Usecase
}

func NewServer(usecase Usecase) *Server {
	return &Server{usecase: usecase}
}

func (s *Server) scoreHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		JumpHistory []int  `json:"jumpHistory"`
		PipeKey     string `json:"pipeKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	score, err := s.usecase.calcScore(req.JumpHistory, req.PipeKey)
	if err != nil {
		http.Error(w, "Failed to get score", http.StatusInternalServerError)
		return
	}
	responseBody := struct {
		Score int `json:"score"`
	}{score}
	if err := json.NewEncoder(w).Encode(responseBody); err != nil {
		http.Error(w, "Failed to encode response body", http.StatusInternalServerError)
		return
	}
}

type Usecase interface {
	calcScore(jumpHistory []int, pipeKey string) (int, error)
}

type ScoreUsecase struct{}

func NewScoreUsecase() Usecase {
	return &ScoreUsecase{}
}

func (u *ScoreUsecase) calcScore(jumpHistory []int, pipeKey string) (int, error) {
	obj := common.NewObject(common.InitialX16, common.InitialY16, 0, pipeKey)

	// Sentinel
	jumpHistory = append(jumpHistory, -1)
	i := 0
	for !obj.Hit() {
		obj.X16 += common.DeltaX16
		if jumpHistory[i] == obj.X16 {
			i++
			obj.Vy16 = -common.VyLimit
		}
		obj.Y16 += obj.Vy16

		// Gravity
		obj.Vy16 += common.DeltaVy16
		if obj.Vy16 > common.VyLimit {
			obj.Vy16 = common.VyLimit
		}
	}
	return obj.Score(), nil
}
