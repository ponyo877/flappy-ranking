package common

import "time"

type Score struct {
	Rank      int
	Name      string
	Score     int
	CreatedAt time.Time
}

func NewScore(rank int, name string, score int, createdAt time.Time) *Score {
	return &Score{
		Rank:      rank,
		Name:      name,
		Score:     score,
		CreatedAt: createdAt,
	}
}
