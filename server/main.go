package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ponyo877/flappy-standings/common"
)

func main() {
	repository := NewScoreRepository(nil)
	usecase := NewScoreUsecase(repository)
	server := NewServer(usecase)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /token", server.tokenHandler)
	mux.HandleFunc("GET /score", server.scoreHandler)
	mux.HandleFunc("POST /score", server.scoreRegisterHandler)
	http.ListenAndServe(":8080", mux)
}

type Server struct {
	usecase Usecase
}

func NewServer(usecase Usecase) *Server {
	return &Server{usecase: usecase}
}

func (s *Server) tokenHandler(w http.ResponseWriter, r *http.Request) {
	token := common.NewUlID()
	pipeKey := common.NewUlID()
	if err := s.usecase.registerSession(token, pipeKey); err != nil {
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

func (s *Server) scoreHandler(w http.ResponseWriter, r *http.Request) {
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

	scores, err := s.usecase.listScore(startDate)
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

func (s *Server) scoreRegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		JumpHistory []int  `json:"jumpHistory"`
		Token       string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	pipeKey, startTime, err := s.usecase.getSession(req.Token)
	if err != nil {
		http.Error(w, "Failed to get pipe key", http.StatusInternalServerError)
		return
	}
	obj, err := s.usecase.getObject(req.JumpHistory, pipeKey)
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

type Usecase interface {
	registerSession(token, pipeKey string) error
	listScore(startDate time.Time) ([]*common.Score, error)
	getSession(token string) (string, time.Time, error)
	getObject(jumpHistory []int, pipeKey string) (*common.Object, error)
}

type ScoreUsecase struct {
	repository Repository
}

func NewScoreUsecase(repository Repository) Usecase {
	return &ScoreUsecase{repository}
}

func (u *ScoreUsecase) getObject(jumpHistory []int, pipeKey string) (*common.Object, error) {
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
	return obj, nil
}

func (u *ScoreUsecase) registerSession(token, pipeKey string) error {
	return u.repository.CreateSession(token, pipeKey)
}

func (u *ScoreUsecase) listScore(startDate time.Time) ([]*common.Score, error) {
	return u.repository.ListScore(startDate)
}

func (u *ScoreUsecase) getSession(token string) (string, time.Time, error) {
	return u.repository.GetSession(token)
}

type Repository interface {
	CreateSession(token, pipeKey string) error
	ListScore(startDate time.Time) ([]*common.Score, error)
	GetSession(token string) (string, time.Time, error)
}

type ScoreRepository struct {
	db *sql.DB
}

func NewScoreRepository(db *sql.DB) Repository {
	return &ScoreRepository{db: db}
}

type Score struct {
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	Score     int       `db:"score"`
	CreatedAt time.Time `db:"created_at"`
}

func (r *ScoreRepository) CreateSession(token, pipeKey string) error {
	query := "INSERT INTO users (token, pipe_key) VALUES (?, ?)"
	if _, err := r.db.Exec(query, token, pipeKey); err != nil {
		return err
	}
	return nil
}

func (r *ScoreRepository) ListScore(startDate time.Time) ([]*common.Score, error) {
	var scoresDB []*Score
	query := "SELECT * FROM scores WHERE created_at >= ? ORDER BY score DESC"
	if err := r.db.QueryRow(query, startDate).Scan(&scoresDB); err != nil {
		return nil, err
	}
	var scores []*common.Score
	for _, score := range scoresDB {
		scores = append(scores, common.NewScore(score.ID, score.Name, score.Score, score.CreatedAt))
	}
	return scores, nil
}

type Session struct {
	ID        int       `db:"id"`
	Token     string    `db:"token"`
	PipeKey   string    `db:"pipe_key"`
	CreatedAt time.Time `db:"created_at"`
}

func (r *ScoreRepository) GetSession(token string) (string, time.Time, error) {
	var session Session
	query := "SELECT * FROM users WHERE token = ?"
	if err := r.db.QueryRow(query, token).Scan(&session); err != nil {
		return "", time.Time{}, err
	}
	return session.PipeKey, session.CreatedAt, nil
}
