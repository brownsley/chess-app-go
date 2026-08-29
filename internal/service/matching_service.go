package service

import (
	"game-server/internal/game"
	"game-server/internal/ws"
	"game-server/utils"
)

type MatchingService struct {
	redisService *RedisService
	chessService *ChessService
	roomManager  *ws.RoomManager
}

func NewMatchingService(redisService *RedisService, chessService *ChessService, rm *ws.RoomManager) *MatchingService {
	return &MatchingService{
		redisService: redisService,
		chessService: chessService,
		roomManager:  rm,
	}
}

func (s *MatchingService) LeaveFromMatchingQueue(modeName game.MatchType, playerId string) error {
	return s.redisService.RemoveFromMatchingQueue(modeName, playerId)
}

func (s *MatchingService) StartMatch(invite ws.InvitePayload) {
	matchId := utils.IdGenerate(8)
	whiteId, blackId := s.resolveInvitePlayers(invite)
	whitePlayer, blackPlayer := s.createPlayers(whiteId, blackId, 0)

	s.chessService.InitializeMatch(matchId, invite.MatchType, whitePlayer, blackPlayer)
}

func (s *MatchingService) ProcessMatch(modeName game.MatchType, playerId string, playerElo int) {
	potentialMatches, err := s.redisService.FindMatchingPlayers(modeName, playerElo, 100)

	if err == nil && len(potentialMatches) > 0 {
		for _, potentialPlayerId := range potentialMatches {
			if potentialPlayerId != playerId {
				matchId := utils.IdGenerate(10)
				whiteId, blackId := s.randomizePlayers(playerId, potentialPlayerId)
				s.removePlayersFromQueue(modeName, playerId, potentialPlayerId)
				whitePlayer, blackPlayer := s.createPlayers(whiteId, blackId, playerElo)
				s.chessService.InitializeMatch(matchId, modeName, whitePlayer, blackPlayer)
				return
			}
		}
	}

	_ = s.redisService.AddToMatchingQueue(modeName, playerId, playerElo)
}

func (s *MatchingService) JoinQueue(modeName game.MatchType, minutes int, playerId string, playerElo int) {
	s.ProcessMatch(modeName, playerId, playerElo)
}
