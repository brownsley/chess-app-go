package models

import "game-server/internal/game"

type JoinQueueRequest struct {
	GameMode  game.MatchType `json:"gameMode"`
	Minutes   int            `json:"minutes"`
	PlayerId  string         `json:"playerId"`
	PlayerElo int            `json:"playerElo"`
}

type LeaveQueueRequest struct {
	GameMode game.MatchType `json:"gameMode"`
	Minutes  int            `json:"minutes"`
	PlayerId string         `json:"playerId"`
}
