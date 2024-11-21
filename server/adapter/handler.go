package adapter

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ponyo877/flappy-standings/common"
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
		http.Error(w, "Failed to encode response body", http.StatusInternalServerError)
		return
	}
}

func (s *Adapter) ListScoreHandler(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	scores, err := s.usecase.ListScore(period)
	if err != nil {
		http.Error(w, "Failed to get score", http.StatusInternalServerError)
		return
	}

	responseBody := struct {
		Scores []ScoreJSON `json:"scores"`
	}{NewScoreJSONList(scores)}
	if err := json.NewEncoder(w).Encode(responseBody); err != nil {
		http.Error(w, "Failed to encode response body", http.StatusInternalServerError)
		return
	}
}

func (s *Adapter) RegisterScoreHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Token       string `json:"token"`
		JumpHistory []int  `json:"jumpHistory"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	score, err := s.usecase.CalcScore(req.JumpHistory, req.Token)
	if err != nil {
		http.Error(w, "Failed to calculate score", http.StatusInternalServerError)
		return
	}
	if err := s.usecase.RegisterScore(req.Name, score); err != nil {
		http.Error(w, "Failed to register score", http.StatusInternalServerError)
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

type ScoreJSON struct {
	Rank      int       `json:"rank"`
	Name      string    `json:"name"`
	Score     int       `json:"score"`
	CreatedAt time.Time `json:"created_at"`
}

func NewScoreJSON(score *common.Score) ScoreJSON {
	return ScoreJSON{
		Rank:      score.Rank,
		Name:      score.Name,
		Score:     score.Score,
		CreatedAt: score.CreatedAt,
	}
}

func NewScoreJSONList(scores []*common.Score) []ScoreJSON {
	scoreJSONs := make([]ScoreJSON, len(scores))
	for i, score := range scores {
		scoreJSONs[i] = NewScoreJSON(score)
	}
	return scoreJSONs
}
