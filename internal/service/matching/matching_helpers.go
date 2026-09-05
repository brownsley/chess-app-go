package service

import (
	"game-server/internal/game"
	"game-server/internal/models"
	"game-server/internal/ws"
	"math/rand/v2"
)

func (s *MatchingService) removePlayersFromQueue(modeName game.MatchType, rawMembers ...string) {
	for _, member := range rawMembers {
		_, _ = s.redisService.RemoveFromMatchingQueue(modeName, member)
	}
}
func (s *MatchingService) createPlayers(whiteId, whiteName, blackId, blackName string, elo int) (models.Player, models.Player) {
	white := models.Player{
		ID:   whiteId,
		Name: whiteName,
		Side: models.White,
		Elo:  elo,
	}
	black := models.Player{
		ID:   blackId,
		Name: blackName,
		Side: models.Black,
		Elo:  elo,
	}
	return white, black
}

func (s *MatchingService) randomizeQueuePlayers(id1, name1, id2, name2 string) (string, string, string, string) {
	if rand.IntN(2) == 0 {
		return id1, name1, id2, name2
	}
	return id2, name2, id1, name1
}

func (s *MatchingService) resolveInvitePlayers(invite ws.InvitePayload) (string, string, string, string) {
	otherId := invite.OtherPlayer
	otherName := invite.OtherPlayerName
	challengerId := invite.ChallengerID
	challengerName := invite.ChallengerName

	switch invite.ColorPreference {
	case ws.ChoiceWhite:
		return challengerId, challengerName, otherId, otherName

	case ws.ChoiceBlack:
		return otherId, otherName, challengerId, challengerName

	case ws.ChoiceRandom:
		if rand.IntN(2) == 0 {
			return challengerId, challengerName, otherId, otherName
		}
		return otherId, otherName, challengerId, challengerName
	default:
		return challengerId, challengerName, otherId, otherName
	}
}
