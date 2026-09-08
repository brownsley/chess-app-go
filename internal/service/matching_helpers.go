package service

import (
	"game-server/db"
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
func (s *MatchingService) createPlayers(whiteId, whiteName, whiteCountry, whiteAvatar string, whteElo int,
	blackId, blackName, blackCountry, blackAvatar string, blackElo int) (models.Player, models.Player) {
	white := models.Player{
		ID:      whiteId,
		Name:    whiteName,
		Avatar:  whiteAvatar,
		Country: whiteCountry,
		Side:    models.White,
		Elo:     whteElo,
	}
	black := models.Player{
		ID:      blackId,
		Name:    blackName,
		Avatar:  blackAvatar,
		Country: blackCountry,
		Side:    models.Black,
		Elo:     blackElo,
	}
	return white, black
}

func (s *MatchingService) randomizeQueuePlayers(id1, name1, id2, name2 string) (string, string, string, string) {
	if rand.IntN(2) == 0 {
		return id1, name1, id2, name2
	}
	return id2, name2, id1, name1
}
func (s *MatchingService) resolveInvitePlayers(invite ws.InvitePayload, receiver *db.User) (models.Player, models.Player) {
	isSenderWhite := false

	switch invite.ColorPreference {
	case ws.ChoiceWhite:
		isSenderWhite = true
	case ws.ChoiceBlack:
		isSenderWhite = false
	case ws.ChoiceRandom:
		isSenderWhite = rand.IntN(2) == 0
	default:
		isSenderWhite = true
	}

	if isSenderWhite {
		return s.createPlayers(
			invite.SenderID, invite.SenderName, invite.SenderCountry, invite.SenderAvatar, invite.SenderElo,
			receiver.UserID, receiver.Name, receiver.Country, receiver.AvatarURL, receiver.Elo)
	}

	return s.createPlayers(
		receiver.UserID, receiver.Name, receiver.Country, receiver.AvatarURL, receiver.Elo,
		invite.SenderID, invite.SenderName, invite.SenderCountry, invite.SenderAvatar, invite.SenderElo,
	)
}
