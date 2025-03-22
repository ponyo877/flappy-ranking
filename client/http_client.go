package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ponyo877/flappy-ranking/common"
)

func (g *Game) fetchToken() {
	resp, err := http.Post(endpoint.JoinPath("api", "tokens").String(), "application/json", nil)
	if err != nil {
		log.Printf("Failed to get token: %v", err)
		return
	}
	defer resp.Body.Close()

	var result struct {
		Token   string `json:"token"`
		PipeKey string `json:"pipeKey"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Failed to decode token response: %v", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to fetch token: %s", resp.Status)
		return
	}

	g.token = result.Token
	g.pipeKey = result.PipeKey
	// log.Printf("Got token: %s, pipeKey: %s", g.token, g.pipeKey)
}

func (g *Game) fetchRanking() {
	g.fetchingRanking = true
	endpoint := endpoint.JoinPath("api", "scores")
	q := endpoint.Query()
	q.Set("period", g.rankingPeriod)
	endpoint.RawQuery = q.Encode()

	resp, err := http.Get(endpoint.String())
	if err != nil {
		log.Printf("Failed to fetch ranking: %v", err)
		g.fetchingRanking = false
		return
	}
	defer resp.Body.Close()

	var result struct {
		Scores []struct {
			Rank        int       `json:"rank"`
			DisplayName string    `json:"display_name"`
			Score       int       `json:"score"`
			CreatedAt   time.Time `json:"created_at"`
		} `json:"scores"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Failed to decode ranking response: %v", err)
		g.fetchingRanking = false
		return
	}

	g.rankings = nil
	for _, s := range result.Scores {
		g.rankings = append(g.rankings, common.NewScore(
			s.Rank, s.DisplayName, s.Score, s.CreatedAt))
	}
	g.fetchingRanking = false
}

func (g *Game) submitScore(playerName string) {
	data := struct {
		DisplayName string `json:"displayName"`
		JumpHistory []int  `json:"jumpHistory"`
	}{
		DisplayName: playerName,
		JumpHistory: g.jumpHistory,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		g.errorMessage = "Error preparing data"
		log.Printf("Failed to marshal score data: %v", err)
		return
	}
	resp, err := http.Post(endpoint.JoinPath("api", "scores", g.token).String(), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		g.errorMessage = "Network error"
		log.Printf("Failed to submit score: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		g.errorMessage = "Server error: " + resp.Status
		log.Printf("Failed to submit score: %s", resp.Status)
		return
	}
	g.scoreSubmitted = true
	// log.Printf("Score submitted successfully")
}

func (g *Game) finishSession() {
	resp, err := http.Post(endpoint.JoinPath("api", "sessions", g.token).String(), "application/json", nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to finish session: %s", resp.Status)
		return
	}
	// log.Printf("Session finished successfully")
}
