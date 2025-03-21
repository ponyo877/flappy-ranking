package adapter

import (
	"time"

	"github.com/ponyo877/flappy-ranking/common"
)

type Usecase interface {
	RegisterScore(displayName string, score int) error
	RegisterSession(token, pipeKey string) error
	ListScore(period string) ([]*common.Score, error)
	CalcScore(jumpHistory []int, token string) (int, error)
	FinishSession(token string) error
}

type Repository interface {
	CreateScore(displayName string, score int) error
	CreateSession(token, pipeKey string) error
	ListScore(startTime time.Time, limit int) ([]*common.Score, error)
	GetSession(token string) (*common.Session, error)
	UpdateSessionFinishedAt(token string) error
}
