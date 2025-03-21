package usecase

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestScoreUsecase_simulateObject(t *testing.T) {
	u := &ScoreUsecase{}

	type args struct {
		jumpHistory []int
		pipeKey     string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "success",
			args: args{
				jumpHistory: []int{736, 1440, 2816, 4928, 6464, 8032, 10432, 11552, 13088, 14880, 15904, 17952, 19392, 20864, 21792, 23200, 24608, 26624, 27968, 29824, 31616, 32896, 34912, 36480, 37664, 39072, 40192, 41824, 43936},
				pipeKey:     "ABCDEFGHIJKLMNOPQRSTUVWXYZ123456",
			},
			want: 9,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := u.simulateObject(tt.args.jumpHistory, tt.args.pipeKey)
			assert.Equal(t, tt.want, got.Score())
		})
	}
}

func TestScoreUsecase_calcStarTime(t *testing.T) {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Date(2024, 11, 16, 1, 2, 3, 4, jst) // Saturday
	u := &ScoreUsecase{}
	type args struct {
		now    time.Time
		period string
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "daily",
			args: args{
				now:    now,
				period: "DAILY",
			},
			want: time.Date(2024, 11, 16, 0, 0, 0, 0, jst),
		},
		{
			name: "weekly",
			args: args{
				now:    now,
				period: "WEEKLY",
			},
			want: time.Date(2024, 11, 10, 0, 0, 0, 0, jst),
		},
		{
			name: "monthly",
			args: args{
				now:    now,
				period: "MONTHLY",
			},
			want: time.Date(2024, 11, 1, 0, 0, 0, 0, jst),
		},
		{
			name: "default",
			args: args{
				now:    now,
				period: "ALL",
			},
			want: time.Time{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := u.calcStarTime(tt.args.now, tt.args.period)
			assert.Equal(t, tt.want, got)
		})
	}
}
