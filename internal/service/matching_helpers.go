package service

import (
	"game-server/internal/game"
	"game-server/internal/models"
	"game-server/internal/ws"
	"math/rand/v2"
)

func (s *MatchingService) removePlayersFromQueue(modeName game.MatchType, playerIds ...string) {
	for _, id := range playerIds {
		_ = s.redisService.RemoveFromMatchingQueue(modeName, id)
	}
}

func (s *MatchingService) createPlayers(whiteId, blackId string, elo int) (models.Player, models.Player) {
	white := models.Player{
		ID:   whiteId,
		Name: "White Player",
		Side: models.White,
		Elo:  elo,
	}
	black := models.Player{
		ID:   blackId,
		Name: "Black Player",
		Side: models.Black,
		Elo:  elo,
	}
	return white, black
}

func (s *MatchingService) randomizePlayers(p1, p2 string) (string, string) {
	players := []string{p1, p2}
	rand.Shuffle(len(players), func(i, j int) {
		players[i], players[j] = players[j], players[i]
	})
	return players[0], players[1]
}

func (s *MatchingService) resolveInvitePlayers(invite ws.InvitePayload) (string, string) {
	if invite.WhitePlayer == invite.ChallengerID {
		return invite.ChallengerID, invite.OtherPlayer
	}
	return invite.OtherPlayer, invite.ChallengerID
}
