package common

import "time"

type Score struct {
	Rank        int
	DisplayName string
	Score       int
	CreatedAt   time.Time
}

func NewScore(rank int, displayName string, score int, createdAt time.Time) *Score {
	return &Score{
		Rank:        rank,
		DisplayName: displayName,
		Score:       score,
		CreatedAt:   createdAt,
	}
}
