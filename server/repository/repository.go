package repository

import (
	"database/sql"
	"time"

	"github.com/ponyo877/flappy-ranking/common"
	"github.com/ponyo877/flappy-ranking/server/adapter"
)

type ScoreRepository struct {
	db *sql.DB
}

func NewScoreRepository(db *sql.DB) adapter.Repository {
	return &ScoreRepository{db: db}
}

type Score struct {
	ID          int    `db:"id"`
	DisplayName string `db:"display_name"`
	Score       int    `db:"score"`
	CreatedAt   uint64 `db:"created_at"`
}

func (r *ScoreRepository) CreateScore(displayName string, score int) error {
	query := "INSERT INTO scores (display_name, score, created_at) VALUES (?, ?, ?)"
	now := time.Now().Unix()
	if _, err := r.db.Exec(query, displayName, score, now); err != nil {
		return err
	}
	return nil
}

func (r *ScoreRepository) CreateSession(token, pipeKey string) error {
	query := "INSERT INTO sessions (token, pipe_key, finished_at, created_at) VALUES (?, ?, ?, ?)"
	now := time.Now().Unix()
	if _, err := r.db.Exec(query, token, pipeKey, now, now); err != nil {
		return err
	}
	return nil
}

func (r *ScoreRepository) ListScore(startDate time.Time, limit int) ([]*common.Score, error) {
	query := "SELECT id, display_name, score, created_at FROM scores WHERE created_at >= ? ORDER BY score DESC LIMIT ?"
	rows, err := r.db.Query(query, startDate.Unix(), limit)
	if err != nil && err != sql.ErrNoRows {
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
		scores = append(scores, common.NewScore(currentRank, s.DisplayName, s.Score, time.Unix(int64(s.CreatedAt), 0)))
		previousScore = s.Score
		rank++
	}

	return scores, nil
}

func (r *ScoreRepository) GetSession(token string) (string, time.Time, error) {
	query := "SELECT pipe_key, created_at FROM sessions WHERE token = ?"
	var pipeKey string
	var createdAt uint64
	if err := r.db.QueryRow(query, token).Scan(&pipeKey, &createdAt); err != nil {
		return "", time.Time{}, err
	}
	return pipeKey, time.Unix(int64(createdAt), 0), nil
}
