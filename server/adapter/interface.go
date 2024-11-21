package adapter

import (
	"time"

	"github.com/ponyo877/flappy-standings/common"
)

type Usecase interface {
	RegisterScore(name string, score int) error
	RegisterSession(token, pipeKey string) error
	ListScore(period string) ([]*common.Score, error)
	CalcScore(jumpHistory []int, token string) (int, error)
}

type Repository interface {
	CreateScore(name string, score int) error
	CreateSession(token, pipeKey string) error
	ListScore(startTime time.Time, limit int) ([]*common.Score, error)
	GetSession(token string) (string, time.Time, error)
}
