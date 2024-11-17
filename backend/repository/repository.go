package repository

import (
	"database/sql"
	"time"

	"github.com/ponyo877/flappy-standings/backend/server"
	"github.com/ponyo877/flappy-standings/common"
)

type ScoreRepository struct {
	db *sql.DB
}

func NewScoreRepository(db *sql.DB) server.Repository {
	return &ScoreRepository{db: db}
}

type Score struct {
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	Score     int       `db:"score"`
	CreatedAt time.Time `db:"created_at"`
}

type Session struct {
	ID        int       `db:"id"`
	Token     string    `db:"token"`
	PipeKey   string    `db:"pipe_key"`
	CreatedAt time.Time `db:"created_at"`
}

func (r *ScoreRepository) CreateScore(name string, score int) error {
	query := "INSERT INTO scores (name, score) VALUES (?, ?)"
	if _, err := r.db.Exec(query, name, score); err != nil {
		return err
	}
	return nil
}

func (r *ScoreRepository) CreateSession(token, pipeKey string) error {
	query := "INSERT INTO users (token, pipe_key) VALUES (?, ?)"
	if _, err := r.db.Exec(query, token, pipeKey); err != nil {
		return err
	}
	return nil
}

func (r *ScoreRepository) ListScore(startDate time.Time, limit int) ([]*common.Score, error) {
	var scoresDB []*Score
	query := "SELECT * FROM scores WHERE created_at >= ? ORDER BY score DESC LIMIT ?"
	if err := r.db.QueryRow(query, startDate, limit).Scan(&scoresDB); err != nil {
		return nil, err
	}
	var scores []*common.Score
	for _, score := range scoresDB {
		scores = append(scores, common.NewScore(score.ID, score.Name, score.Score, score.CreatedAt))
	}
	return scores, nil
}

func (r *ScoreRepository) GetSession(token string) (string, time.Time, error) {
	var session Session
	query := "SELECT * FROM users WHERE token = ?"
	if err := r.db.QueryRow(query, token).Scan(&session); err != nil {
		return "", time.Time{}, err
	}
	return session.PipeKey, session.CreatedAt, nil
}
