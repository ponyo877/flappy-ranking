package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ponyo877/flappy-standings/common"
)

type Server struct {
	usecase Usecase
}

func NewServer(usecase Usecase) *Server {
	return &Server{usecase: usecase}
}

func (s *Server) TokenHandler(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) ScoreHandler(w http.ResponseWriter, r *http.Request) {
	period := "DAILY"
	var startDate time.Time
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		http.Error(w, "Failed to load location", http.StatusInternalServerError)
		return
	}
	todayHead := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, jst)
	switch period {
	case "DAILY":
		startDate = todayHead
	case "WEEKLY":
		offset := int(time.Now().Weekday()) - int(time.Sunday)
		startDate = todayHead.AddDate(0, 0, -offset).In(jst)
	case "MONTHLY":
		startDate = time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, jst)
	default:
		startDate = time.Time{}
	}
	limit := 20
	scores, err := s.usecase.ListScore(startDate, limit)
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

func (s *Server) ScoreRegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		JumpHistory []int  `json:"jumpHistory"`
		Token       string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	pipeKey, startTime, err := s.usecase.GetSession(req.Token)
	if err != nil {
		http.Error(w, "Failed to get pipe key", http.StatusInternalServerError)
		return
	}
	obj, err := s.usecase.GetObject(req.JumpHistory, pipeKey)
	if err != nil {
		http.Error(w, "Failed to calculate score", http.StatusInternalServerError)
		return
	}
	if !obj.IsValidTimeDiff(startTime, time.Now()) {
		http.Error(w, "Invalid game end time", http.StatusBadRequest)
		return
	}
	responseBody := struct {
		Score int `json:"score"`
	}{obj.Score()}
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
