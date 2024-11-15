package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ponyo877/flappy-standings/common"
)

func main() {
	server := NewServer(NewScoreUsecase())
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello", server.helloHandler)
	mux.HandleFunc("POST /score", server.scoreHandler)
	http.ListenAndServe(":8080", mux)
}

type Server struct {
	usecase Usecase
}

func NewServer(usecase Usecase) *Server {
	return &Server{usecase: usecase}
}

func (s *Server) helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
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
	score, err := s.usecase.getScore(req.JumpHistory, req.PipeKey)
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
	getScore(jumpHistory []int, pipeKey string) (int, error)
}

type ScoreUsecase struct{}

func NewScoreUsecase() Usecase {
	return &ScoreUsecase{}
}

// curl -X POST http://localhost:8080/score -H "Content-Type: application/json" -d '{"jumpHistory":[1248,2400,3488,5056,6400,7872,9984,11488,12992,14688,16000,17984,19520,20960,22304,23200,24640],"pipeKey":"ABCDEFGHIJKLMNOPQRSTUVWXYZ123456"}' | jq .
func (u *ScoreUsecase) getScore(jumpHistory []int, pipeKey string) (int, error) {
	obj := common.NewObject(common.InitialX16, common.InitialY16, 0, pipeKey)
	// isHit := false
	// g.obj.X16 += common.DeltaX16
	// g.cameraX += common.DeltaCameraX
	// if g.isKeyJustPressed() {
	// 	g.jumpHistory = append(g.jumpHistory, g.obj.X16)
	// 	g.obj.Vy16 = -common.VyLimit
	// 	if err := g.jumpPlayer.Rewind(); err != nil {
	// 		return err
	// 	}
	// 	g.jumpPlayer.Play()
	// }
	// g.obj.Y16 += g.obj.Vy16

	// // Gravity
	// g.obj.Vy16 += common.DeltaVy16
	// if g.obj.Vy16 > common.VyLimit {
	// 	g.obj.Vy16 = common.VyLimit
	// }

	// if g.obj.Hit() {
	// 	log.Printf("debug jumpHistory: %v", g.jumpHistory)
	// 	g.jumpHistory = []int{}
	// 	if err := g.hitPlayer.Rewind(); err != nil {
	// 		return err
	// 	}
	// 	g.hitPlayer.Play()
	// 	g.mode = ModeGameOver
	// 	g.gameoverCount = 30
	// }
	i := 0
	for {
		obj.X16 += common.DeltaX16
		if jumpHistory[i] == obj.X16 {
			i++
			obj.Vy16 = -common.VyLimit
		}
		obj.Y16 += obj.Vy16
		obj.Vy16 += common.DeltaVy16
		if obj.Vy16 > common.VyLimit {
			obj.Vy16 = common.VyLimit
		}
		if obj.Hit() {
			break
		}
	}
	// for _, j := range jumpHistory {
	// 	for obj.X16 < j {
	// 		if obj.Hit() {
	// 			isHit = true
	// 			break
	// 		}
	// 		obj.X16 += common.DeltaX16
	// 		obj.Y16 += obj.Vy16
	// 		// Gravity
	// 		obj.Vy16 += common.DeltaVy16
	// 		if obj.Vy16 > common.VyLimit {
	// 			obj.Vy16 = common.VyLimit
	// 		}
	// 	}
	// 	if isHit {
	// 		break
	// 	}
	// 	obj.Vy16 = -common.VyLimit
	// }
	return obj.Score(), nil
}
