package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScoreUsecase_calcScore(t *testing.T) {
	u := &ScoreUsecase{}

	type args struct {
		jumpHistory []int
		pipeKey     string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		errFunc func(t assert.TestingT, err error)
	}{
		{
			name: "success",
			args: args{
				jumpHistory: []int{736, 1440, 2816, 4928, 6464, 8032, 10432, 11552, 13088, 14880, 15904, 17952, 19392, 20864, 21792, 23200, 24608, 26624, 27968, 29824, 31616, 32896, 34912, 36480, 37664, 39072, 40192, 41824, 43936},
				pipeKey:     "ABCDEFGHIJKLMNOPQRSTUVWXYZ123456",
			},
			want: 9,
			errFunc: func(t assert.TestingT, err error) {
				assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := u.calcScore(tt.args.jumpHistory, tt.args.pipeKey)
			assert.Equal(t, tt.want, got)
			tt.errFunc(t, err)
		})
	}
}
