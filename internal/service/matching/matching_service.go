package service

import (
	"fmt"
	"game-server/internal/game"

	chess "game-server/internal/service/chess"
	redis "game-server/internal/service/redis"

	"game-server/internal/ws"
	"game-server/utils"
	"strings"
)

type MatchingService struct {
	redisService *redis.RedisService
	chessService *chess.ChessService
	roomManager  *ws.RoomManager
}

func NewMatchingService(redisService *redis.RedisService, chessService *chess.ChessService, rm *ws.RoomManager) *MatchingService {
	return &MatchingService{
		redisService: redisService,
		chessService: chessService,
		roomManager:  rm,
	}
}

func (s *MatchingService) LeaveFromMatchingQueue(modeName game.MatchType, playerId string, playerName string) error {
	rawMember := fmt.Sprintf("%s:%s", playerId, playerName)
	_, err := s.redisService.RemoveFromMatchingQueue(modeName, rawMember)
	return err
}

func (s *MatchingService) StartFriendMatch(invite ws.InvitePayload, withFriend bool) {
	matchId := utils.IdGenerate(8, withFriend)
	whiteId, whiteName, blackId, blackName := s.resolveInvitePlayers(invite)
	whitePlayer, blackPlayer := s.createPlayers(whiteId, whiteName, blackId, blackName, 0)
	s.chessService.InitializeMatch(matchId, invite.MatchType, whitePlayer, blackPlayer)
}

func (s *MatchingService) ProcessQueueMatch(modeName game.MatchType, playerId string, playerName string, playerElo int) {
	potentialMatches, err := s.redisService.FindMatchingPlayers(modeName, playerElo, 100)

	if err == nil && len(potentialMatches) > 0 {
		for _, rawMember := range potentialMatches {
			parts := strings.SplitN(rawMember, ":", 2)
			if len(parts) != 2 {
				continue
			}
			potentialPlayerId := parts[0]
			potentialPlayerName := parts[1]

			if potentialPlayerId != playerId {
				removed, err := s.redisService.RemoveFromMatchingQueue(modeName, rawMember)
				if err != nil || !removed {
					continue
				}

				currentRawMember := fmt.Sprintf("%s:%s", playerId, playerName)
				_, _ = s.redisService.RemoveFromMatchingQueue(modeName, currentRawMember)

				matchId := utils.IdGenerate(10, false)

				whiteId, whiteName, blackId, blackName := s.randomizeQueuePlayers(
					playerId, playerName,
					potentialPlayerId, potentialPlayerName,
				)

				whitePlayer, blackPlayer := s.createPlayers(whiteId, whiteName, blackId, blackName, playerElo)
				s.chessService.InitializeMatch(matchId, modeName, whitePlayer, blackPlayer)
				return
			}
		}
	}

	_ = s.redisService.AddToMatchingQueue(modeName, playerId, playerName, playerElo)
}

func (s *MatchingService) JoinQueue(modeName game.MatchType, minutes int, playerId string, playerName string, playerElo int) {
	s.ProcessQueueMatch(modeName, playerId, playerName, playerElo)
}
