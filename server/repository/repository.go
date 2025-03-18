package repository

import (
	"database/sql"
	"time"

	"github.com/ponyo877/flappy-standings/common"
	"github.com/ponyo877/flappy-standings/server/adapter"
)

type ScoreRepository struct {
	db *sql.DB
}

func NewScoreRepository(db *sql.DB) adapter.Repository {
	return &ScoreRepository{db: db}
}

type Score struct {
	ID          int       `db:"id"`
	DisplayName string    `db:"display_name"`
	Score       int       `db:"score"`
	CreatedAt   time.Time `db:"created_at"`
}

type Session struct {
	ID        int       `db:"id"`
	Token     string    `db:"token"`
	PipeKey   string    `db:"pipe_key"`
	CreatedAt time.Time `db:"created_at"`
}

func (r *ScoreRepository) CreateScore(displayName string, score int) error {
	query := "INSERT INTO scores (display_name, score) VALUES (?, ?)"
	if _, err := r.db.Exec(query, displayName, score); err != nil {
		return err
	}
	return nil
}

func (r *ScoreRepository) CreateSession(token, pipeKey string) error {
	query := "INSERT INTO sessions (token, pipe_key) VALUES (?, ?)"
	if _, err := r.db.Exec(query, token, pipeKey); err != nil {
		return err
	}
	return nil
}

func (r *ScoreRepository) ListScore(startDate time.Time, limit int) ([]*common.Score, error) {
	query := "SELECT id, display_name, score, created_at FROM scores WHERE created_at >= ? ORDER BY score DESC LIMIT ?"
	rows, err := r.db.Query(query, startDate, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []*common.Score
	rank := 1
	previousScore := -1
	previousRank := 0

	for rows.Next() {
		var s Score
		if err := rows.Scan(&s.ID, &s.DisplayName, &s.Score, &s.CreatedAt); err != nil {
			return nil, err
		}
		currentRank := previousRank
		if s.Score != previousScore {
			currentRank = rank
			previousRank = rank
		}
		scores = append(scores, common.NewScore(currentRank, s.DisplayName, s.Score, s.CreatedAt))
		previousScore = s.Score
		rank++
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return scores, nil
}

func (r *ScoreRepository) GetSession(token string) (string, time.Time, error) {
	query := "SELECT pipe_key, created_at FROM sessions WHERE token = ?"
	var pipeKey string
	var createdAt time.Time
	if err := r.db.QueryRow(query, token).Scan(&pipeKey, &createdAt); err != nil {
		return "", time.Time{}, err
	}
	return pipeKey, createdAt, nil
}
