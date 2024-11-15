package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ponyo877/flappy-standings/common"
)

func main() {
	server := &Server{}
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", server.helloHandler)
	mux.HandleFunc("/score", server.scoreHandler)
	http.ListenAndServe(":8080", mux)
}

type Server struct {
	usecase Usecase
}

func (s *Server) helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

func (s *Server) scoreHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		JumpHistory []int `json:"jumpHistory"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	jumpHistory := requestBody.JumpHistory
	score, err := s.usecase.getScore(jumpHistory)
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
	getScore(jumpHistory []int) (int, error)
}

type ScoreUsecase struct{}

func NewScoreUsecase() Usecase {
	return &ScoreUsecase{}
}

func (u *ScoreUsecase) getScore(jumpHistory []int) (int, error) {
	x16 := common.InitialX16
	y16 := common.InitialY16
	vy16 := 0
	x16 += common.DeltaX16
	isHit := false
	for _, j := range jumpHistory {
		for x16 < j {
			if hit() {
				isHit = true
				break
			}
			x16 += common.DeltaX16
			y16 += vy16
			// Gravity
			vy16 += common.DeltaVy16
			if vy16 > common.VyLimit {
				vy16 = common.VyLimit
			}
		}
		if isHit {
			break
		}
		vy16 = -common.VyLimit
	}
	return score(x16), nil
}

func hit() bool {
	return false
}

func score(x16 int) int {
	x := common.FloorDiv(x16, common.Unit) / common.TileSize
	if (x - common.PipeStartOffsetX) <= 0 {
		return 0
	}
	return common.FloorDiv(x-common.PipeStartOffsetX, common.PipeIntervalX)
}
