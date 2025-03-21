package common

import "time"

type Session struct {
	Token      string
	PipeKey    string
	FinishedAt time.Time
	CreatedAt  time.Time
}

func NewSession(token, pipeKey string, finishedAt, createdAt time.Time) *Session {
	return &Session{
		Token:      token,
		PipeKey:    pipeKey,
		FinishedAt: finishedAt,
		CreatedAt:  createdAt,
	}
}
