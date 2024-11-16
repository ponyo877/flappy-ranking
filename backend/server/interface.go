package server

import (
	"time"

	"github.com/ponyo877/flappy-standings/common"
)

type Usecase interface {
	RegisterSession(token, pipeKey string) error
	ListScore(startDate time.Time, limit int) ([]*common.Score, error)
	GetSession(token string) (string, time.Time, error)
	GetObject(jumpHistory []int, pipeKey string) (*common.Object, error)
}

type Repository interface {
	CreateSession(token, pipeKey string) error
	ListScore(startDate time.Time, limit int) ([]*common.Score, error)
	GetSession(token string) (string, time.Time, error)
}
