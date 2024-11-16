package usecase

import (
	"time"

	"github.com/ponyo877/flappy-standings/backend/server"
	"github.com/ponyo877/flappy-standings/common"
)

type ScoreUsecase struct {
	repository server.Repository
}

func NewScoreUsecase(repository server.Repository) server.Usecase {
	return &ScoreUsecase{repository}
}

func (u *ScoreUsecase) GetObject(jumpHistory []int, pipeKey string) (*common.Object, error) {
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

func (u *ScoreUsecase) RegisterSession(token, pipeKey string) error {
	return u.repository.CreateSession(token, pipeKey)
}

func (u *ScoreUsecase) ListScore(startDate time.Time, limit int) ([]*common.Score, error) {
	return u.repository.ListScore(startDate, limit)
}

func (u *ScoreUsecase) GetSession(token string) (string, time.Time, error) {
	return u.repository.GetSession(token)
}
