package service

import (
	"log"
	"sync"
	"time"

	"game-server/internal/enum"
	"game-server/internal/game"
	chessgame "game-server/internal/game"
	"game-server/internal/models"
	"game-server/internal/ws"

	"github.com/corentings/chess/v2"
)

type ChessService struct {
	redisService *RedisService
	gameService  *GameService
	roomManager  *ws.RoomManager
	matchTimers  sync.Map
}

func NewChessService(redisService *RedisService, gameService *GameService, rm *ws.RoomManager) *ChessService {
	return &ChessService{
		redisService: redisService,
		gameService:  gameService,
		roomManager:  rm,
	}
}

func (s *ChessService) InitializeMatch(matchId string, matchType game.MatchType, white, black models.Player) {
	g := chess.NewGame()
	fen := g.FEN()
	_ = s.redisService.SetBoardFen(matchId, fen)
	_ = s.redisService.SetMatchType(matchId, string(matchType))

	timeConfig := matchType.GetTimeConfig()
	timer := chessgame.NewMatchTimer(timeConfig, func(isWhite bool) {
		s.HandleTimeout(matchId, isWhite, white, black)
	})
	s.matchTimers.Store(matchId, timer)
	timer.StartWhite()

	if wJson, bJson, ok := s.marshalPlayers(white, black); ok {
		_ = s.redisService.SetPlayers(matchId, map[string]string{"WHITE": wJson, "BLACK": bJson})
	}

	maxMoveTimeMs := timeConfig.MaxMoveTime.Milliseconds()
	s.roomManager.SendMatchFoundToBoth(matchId, white.ID, black.ID, s.createMatchFoundPayload(matchId, white, black, fen, maxMoveTimeMs))

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-timer.StopChan():
				return
			case <-ticker.C:
				val, exists := s.matchTimers.Load(matchId)
				if !exists {
					return
				}
				mt, ok := val.(*chessgame.MatchTimer)
				if !ok {
					return
				}

				wTime := mt.WhiteTimer.GetRemaining().Milliseconds()
				bTime := mt.BlackTimer.GetRemaining().Milliseconds()

				currentFen, _ := s.redisService.GetBoardFen(matchId)
				moves, _ := s.redisService.GetMoves(matchId)

				opt, _ := chess.FEN(currentFen)
				currentGame := chess.NewGame(opt)
				nextPlayerID := s.getExpectedPlayerID(currentGame, white, black)

				s.roomManager.SendMatchMoveProcess(
					matchId,
					s.createGameStatePayload(matchId, currentFen, nextPlayerID, moves, wTime, bTime, maxMoveTimeMs),
				)
			}
		}
	}()
}

func (s *ChessService) HandleTimeout(matchId string, isWhiteTimeout bool, white, black models.Player) {
	winnerId := black.ID
	if !isWhiteTimeout {
		winnerId = white.ID
	}

	_ = s.gameService.ProcessGameResult(white.ID, white.Elo, black.ID, black.Elo, winnerId, false)

	s.roomManager.SendMatchComplete(matchId, s.createMatchCompletePayload(matchId, winnerId, false, enum.ReasonTimeOut))
	s.cleanupMatch(matchId)
}

func (s *ChessService) ProcessMove(movePayload ws.MovePayload) {
	matchId := movePayload.MatchId
	fen, err := s.redisService.GetBoardFen(matchId)
	if err != nil || fen == "" {
		return
	}

	matchTypeStr, err := s.redisService.GetMatchType(matchId)
	var maxMoveTimeMs int64 = 0
	if err == nil {
		matchType := game.MatchType(matchTypeStr)
		maxMoveTimeMs = matchType.GetTimeConfig().MaxMoveTime.Milliseconds()
	}

	white, black, ok := s.getMatchPlayers(matchId)
	if !ok {
		return
	}

	opt, err := chess.FEN(fen)
	if err != nil {
		return
	}
	g := chess.NewGame(opt)

	expectedID := s.getExpectedPlayerID(g, white, black)
	if expectedID == "" || expectedID != movePayload.PlayerID {
		return
	}

	uci := s.parseUCI(movePayload)

	if err := g.PushNotationMove(uci, chess.UCINotation{}, nil); err != nil {
		return
	}

	newFen := g.FEN()
	_ = s.redisService.SetBoardFen(matchId, newFen)
	_ = s.redisService.PushMove(matchId, uci)

	wTime, bTime := s.updateAndGetTimers(matchId)
	moves, _ := s.redisService.GetMoves(matchId)

	reason, isDraw := s.determineGameReason(g.Position().Status(), g.Outcome() != chess.NoOutcome)
	winner := s.determineWinner(g.Outcome(), white, black, isDraw)

	if g.Outcome() != chess.NoOutcome {
		var winnerID string
		if winner != nil {
			winnerID = winner.ID
		}

		err := s.gameService.ProcessGameResult(
			white.ID, white.Elo,
			black.ID, black.Elo,
			winnerID,
			isDraw,
		)
		if err != nil {
			log.Printf("[GAME ERROR] ProcessGameResult failed: %v", err)
		}

		s.roomManager.SendMatchComplete(matchId, s.createMatchCompletePayload(matchId, winnerID, isDraw, reason))
		s.cleanupMatch(matchId)
		return
	}

	s.roomManager.SendMatchMoveProcess(
		matchId,
		s.createGameStatePayload(matchId, newFen, s.getExpectedPlayerID(g, white, black), moves, wTime, bTime, maxMoveTimeMs),
	)
}

func (s *ChessService) ProcessResign(matchId string, resignPayload ws.ResignPayload) {
	white, black, ok := s.getMatchPlayers(matchId)
	if !ok {
		return
	}

	winnerId, valid := s.determineResignWinner(resignPayload.PlayerID, white, black)
	if !valid {
		return
	}

	_ = s.gameService.ProcessGameResult(white.ID, white.Elo, black.ID, black.Elo, winnerId, false)

	payload := s.createMatchCompletePayload(matchId, winnerId, false, enum.ReasonResignation)
	s.roomManager.SendMatchComplete(matchId, payload)
	s.cleanupMatch(matchId)
}

func (s *ChessService) ProcessOfferDraw(matchId string, offerDraw ws.OfferDrawPayload) {
	_, _, ok := s.getMatchPlayers(matchId)
	if !ok {
		return
	}
	s.roomManager.SendOfferDraw(matchId, offerDraw)
}

func (s *ChessService) ProcessAcceptDraw(matchId string, acceptDraw ws.AcceptDrawPayload) {
	white, black, ok := s.getMatchPlayers(matchId)
	if !ok {
		return
	}

	err := s.gameService.ProcessGameResult(
		white.ID, white.Elo,
		black.ID, black.Elo,
		"",
		true,
	)
	if err != nil {
		log.Printf("[GAME ERROR] ProcessGameResult for Draw failed: %v", err)
	}

	s.roomManager.SendMatchComplete(matchId, s.createMatchCompletePayload(
		matchId,
		"",
		true,
		enum.ReasonDrawAgreement,
	))

	s.cleanupMatch(matchId)
}

func (s *ChessService) ProcessDeclineDraw(matchId string, declineDraw ws.DeclineDrawPayload) {
	_, _, ok := s.getMatchPlayers(matchId)
	if !ok {
		return
	}
	s.roomManager.SendDeclineDraw(matchId, declineDraw)
}
