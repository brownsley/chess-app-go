package service

import (
	"fmt"
	"game-server/internal/game"
	"game-server/internal/request"

	"game-server/internal/ws"
	"game-server/utils"
	"strings"
)

type MatchingService struct {
	redisService *RedisService
	chessService *ChessService
	userService  *UserService
	roomManager  *ws.RoomManager
}

func NewMatchingService(redisService *RedisService, chessService *ChessService, userService *UserService, rm *ws.RoomManager) *MatchingService {
	return &MatchingService{
		redisService: redisService,
		chessService: chessService,
		userService:  userService,
		roomManager:  rm,
	}
}

func (s *MatchingService) LeaveFromMatchingQueue(modeName game.MatchType, playerId string, playerName string) error {
	rawMember := fmt.Sprintf("%s:%s", playerId, playerName)
	_, err := s.redisService.RemoveFromMatchingQueue(modeName, rawMember)
	return err
}

func (s *MatchingService) StartFriendMatch(invite ws.InvitePayload, withFriend bool) error {
	matchId := utils.IdGenerate(8, withFriend)

	receiver, err := s.userService.GetUserByUserID(invite.ReceiverID)
	if err != nil {
		return err
	}

	whitePlayer, blackPlayer := s.resolveInvitePlayers(invite, receiver)

	s.chessService.InitializeMatch(matchId, invite.MatchType, whitePlayer, blackPlayer)
	return nil
}

func (s *MatchingService) ProcessQueueMatch(modeName game.MatchType, req request.JoinQueueRequest) {
	potentialMatches, err := s.redisService.FindMatchingPlayers(req.GameMode, req.PlayerElo, 120)

	if err == nil && len(potentialMatches) > 0 {
		for _, z := range potentialMatches {
			rawMember := z.Member.(string)
			potentialElo := int(z.Score)

			parts := strings.SplitN(rawMember, ":", 4)
			if len(parts) != 4 {
				continue
			}

			potentialPlayerId := parts[0]
			potentialPlayerName := parts[1]
			potentialCountry := parts[2]
			potentialAvatarURL := parts[3]

			if potentialPlayerId != req.PlayerId {
				removed, err := s.redisService.RemoveFromMatchingQueue(req.GameMode, rawMember)
				if err != nil || !removed {
					continue
				}

				currentRawMember := fmt.Sprintf("%s:%s:%s:%s", req.PlayerId, req.PlayerName, req.Country, req.AvatarURL)
				_, _ = s.redisService.RemoveFromMatchingQueue(req.GameMode, currentRawMember)

				matchId := utils.IdGenerate(10, false)

				whiteId, whiteName, blackId, blackName := s.randomizeQueuePlayers(
					req.PlayerId, req.PlayerName,
					potentialPlayerId, potentialPlayerName,
				)

				var whiteCountry, whiteAvatar, blackCountry, blackAvatar string
				var whiteElo, blackElo int

				if whiteId == req.PlayerId {
					whiteCountry, whiteAvatar, whiteElo = req.Country, req.AvatarURL, req.PlayerElo
					blackCountry, blackAvatar, blackElo = potentialCountry, potentialAvatarURL, potentialElo
				} else {
					whiteCountry, whiteAvatar, whiteElo = potentialCountry, potentialAvatarURL, potentialElo
					blackCountry, blackAvatar, blackElo = req.Country, req.AvatarURL, req.PlayerElo
				}

				whitePlayer, blackPlayer := s.createPlayers(
					whiteId, whiteName, whiteCountry, whiteAvatar, whiteElo,
					blackId, blackName, blackCountry, blackAvatar, blackElo,
				)

				s.chessService.InitializeMatch(matchId, req.GameMode, whitePlayer, blackPlayer)
				return
			}
		}
	}

	_ = s.redisService.AddToMatchingQueue(req.GameMode, req)
}
func (s *MatchingService) JoinQueue(modeName game.MatchType, req request.JoinQueueRequest) {
	s.ProcessQueueMatch(modeName, req)
}
