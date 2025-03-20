package usecase

import (
	"fmt"
	"time"

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
	startTime := u.calcStarTime(time.Now(), period)
	return u.repository.ListScore(startTime, limit)
}

func (u *ScoreUsecase) calcStarTime(now time.Time, period string) time.Time {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	todayHead := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, jst)
	weekdayDiff := int(now.Weekday()) - int(time.Sunday)
	switch period {
	case "DAILY":
		return todayHead
	case "WEEKLY":
		return todayHead.AddDate(0, 0, -weekdayDiff).In(jst)
	case "MONTHLY":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, jst)
	default:
		return time.Time{}
	}
}

func (u *ScoreUsecase) CalcScore(jumpHistory []int, token string) (int, error) {
	pipeKey, startTime, err := u.repository.GetSession(token)
	if err != nil {
		return 0, err
	}
	obj := u.simulateObject(jumpHistory, pipeKey)
	endTime := time.Now()

	// Validate Play Time
	if !obj.IsValidTimeDiff(startTime, endTime) {
		return 0, fmt.Errorf("invalid end time: startTime=%v, endTime=%v", startTime, endTime)
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
