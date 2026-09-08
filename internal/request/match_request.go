package request

import "game-server/internal/game"

type JoinQueueRequest struct {
	GameMode   game.MatchType `json:"gameMode"`
	Minutes    int            `json:"minutes"`
	PlayerId   string         `json:"playerId"`
	PlayerName string         `json:"playerName"`
	AvatarURL  string         `json:"avatarUrl"`
	PlayerElo  int            `json:"playerElo"`
	Country    string         `json:"country"`
}

type LeaveQueueRequest struct {
	GameMode   game.MatchType `json:"gameMode"`
	Minutes    int            `json:"minutes"`
	PlayerId   string         `json:"playerId"`
	PlayerName string         `json:"playerName"`
}
