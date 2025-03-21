package adapter

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ponyo877/flappy-ranking/common"
)

type Adapter struct {
	usecase Usecase
}

func NewAdapter(usecase Usecase) *Adapter {
	return &Adapter{usecase: usecase}
}

func (s *Adapter) GenerateTokenHandler(w http.ResponseWriter, r *http.Request) {
	token := common.NewUlID()
	pipeKey := common.NewUlID()
	if err := s.usecase.RegisterSession(token, pipeKey); err != nil {
		log.Printf("Failed to register session: %v", err)
		http.Error(w, "Failed to register session", http.StatusInternalServerError)
		return
	}
	responseBody := struct {
		Token   string `json:"token"`
		PipeKey string `json:"pipeKey"`
	}{
		Token:   token,
		PipeKey: pipeKey,
	}
	if err := json.NewEncoder(w).Encode(responseBody); err != nil {
		log.Printf("Failed to encode response body: %v", err)
		http.Error(w, "Failed to encode response body", http.StatusInternalServerError)
		return
	}
}

func (s *Adapter) ListScoreHandler(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	scores, err := s.usecase.ListScore(period)
	if err != nil {
		log.Printf("Failed to get score: %v", err)
		http.Error(w, "Failed to get score", http.StatusInternalServerError)
		return
	}

	responseBody := struct {
		Scores []ScoreJSON `json:"scores"`
	}{NewScoreJSONList(scores)}
	if err := json.NewEncoder(w).Encode(responseBody); err != nil {
		log.Printf("Failed to encode response body: %v", err)
		http.Error(w, "Failed to encode response body", http.StatusInternalServerError)
		return
	}
}

func (s *Adapter) RegisterScoreHandler(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		log.Printf("Token not provided in path")
		http.Error(w, "Token not provided", http.StatusBadRequest)
		return
	}
	var req struct {
		DisplayName string `json:"displayName"`
		JumpHistory []int  `json:"jumpHistory"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	score, err := s.usecase.CalcScore(req.JumpHistory, token)
	if err != nil {
		log.Printf("Failed to calculate score: %v", err)
		http.Error(w, "Failed to calculate score", http.StatusBadRequest)
		return
	}
	if err := s.usecase.RegisterScore(req.DisplayName, score); err != nil {
		log.Printf("Failed to register score: %v", err)
		http.Error(w, "Failed to register score", http.StatusInternalServerError)
		return
	}
	responseBody := struct {
		Score int `json:"score"`
	}{score}
	if err := json.NewEncoder(w).Encode(responseBody); err != nil {
		log.Printf("Failed to encode response body: %v", err)
		http.Error(w, "Failed to encode response body", http.StatusInternalServerError)
		return
	}
}

func (s *Adapter) FinishSessionHandler(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		log.Printf("Token not provided in path")
		http.Error(w, "Token not provided", http.StatusBadRequest)
		return
	}
	if err := s.usecase.FinishSession(token); err != nil {
		log.Printf("Failed to finish session: %v", err)
		http.Error(w, "Failed to finish session", http.StatusInternalServerError)
		return
	}
	responseBody := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}
	if err := json.NewEncoder(w).Encode(responseBody); err != nil {
		log.Printf("Failed to encode response body: %v", err)
		http.Error(w, "Failed to encode response body", http.StatusInternalServerError)
		return
	}
}

type ScoreJSON struct {
	Rank        int       `json:"rank"`
	DisplayName string    `json:"display_name"`
	Score       int       `json:"score"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewScoreJSON(score *common.Score) ScoreJSON {
	return ScoreJSON{
		Rank:        score.Rank,
		DisplayName: score.DisplayName,
		Score:       score.Score,
		CreatedAt:   score.CreatedAt,
	}
}

func NewScoreJSONList(scores []*common.Score) []ScoreJSON {
	scoreJSONs := make([]ScoreJSON, len(scores))
	for i, score := range scores {
		scoreJSONs[i] = NewScoreJSON(score)
	}
	return scoreJSONs
}
