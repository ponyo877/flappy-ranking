package usecase

import (
	"fmt"
	"time"
	_ "time/tzdata" // https://github.com/golang/go/issues/44408

	"github.com/ponyo877/flappy-ranking/common"
	"github.com/ponyo877/flappy-ranking/server/adapter"
)

type ScoreUsecase struct {
	repository adapter.Repository
}

func NewScoreUsecase(repository adapter.Repository) adapter.Usecase {
	return &ScoreUsecase{repository}
}

func (u *ScoreUsecase) RegisterScore(name string, score int) error {
	return u.repository.CreateScore(name, score)
}

func (u *ScoreUsecase) RegisterSession(token, pipeKey string) error {
	return u.repository.CreateSession(token, pipeKey)
}

func (u *ScoreUsecase) ListScore(period string) ([]*common.Score, error) {
	limit := 10
	startTime, err := u.calcStarTime(time.Now(), period)
	if err != nil {
		return nil, err
	}
	return u.repository.ListScore(startTime, limit)
}

func (u *ScoreUsecase) calcStarTime(now time.Time, period string) (time.Time, error) {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.Time{}, err
	}
	todayHead := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, jst)
	weekdayDiff := int(now.Weekday()) - int(time.Sunday)
	switch period {
	case "DAILY":
		return todayHead, nil
	case "WEEKLY":
		return todayHead.AddDate(0, 0, -weekdayDiff).In(jst), nil
	case "MONTHLY":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, jst), nil
	default:
		return time.Time{}, nil
	}
}

func (u *ScoreUsecase) CalcScore(jumpHistory []int, token string) (int, error) {
	s, err := u.repository.GetSession(token)
	if err != nil {
		return 0, err
	}
	obj := u.simulateObject(jumpHistory, s.PipeKey)

	// Validate Play Time
	if !obj.IsValidTimeDiff(s.CreatedAt, s.FinishedAt) {
		return 0, fmt.Errorf("invalid end time: startTime=%v, endTime=%v", s.CreatedAt, s.FinishedAt)
	}
	return obj.Score(), nil
}

func (u *ScoreUsecase) simulateObject(jumpHistory []int, pipeKey string) *common.Object {
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
	return obj
}

func (u *ScoreUsecase) FinishSession(token string) error {
	return u.repository.UpdateSessionFinishedAt(token)
}
