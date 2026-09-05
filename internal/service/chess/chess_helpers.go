package service

import (
	"encoding/json"
	"game-server/internal/enum"
	chessgame "game-server/internal/game"
	"game-server/internal/models"
	"game-server/internal/ws"
	"strings"

	"github.com/corentings/chess/v2"
)

func (s *ChessService) getExpectedPlayerID(g *chess.Game, white, black models.Player) string {
	if g.Position().Turn() == chess.White {
		return white.ID
	}
	return black.ID
}

func (s *ChessService) determineResignWinner(resigningPlayerID string, white, black models.Player) (string, bool) {
	if resigningPlayerID == black.ID {
		return white.ID, true
	}
	if resigningPlayerID == white.ID {
		return black.ID, true
	}
	return "", false
}

func (s *ChessService) determineWinner(outcome chess.Outcome, white, black models.Player, isDraw bool) string {
	if isDraw {
		return ""
	}
	if outcome == chess.WhiteWon {
		return white.ID
	}
	if outcome == chess.BlackWon {
		return black.ID
	}
	return ""
}

func (s *ChessService) parseUCI(movePayload ws.MovePayload) string {
	return strings.ToLower(movePayload.From + movePayload.To + movePayload.Promotion)
}

func (s *ChessService) marshalPlayers(white, black models.Player) (string, string, bool) {
	wJson, err1 := json.Marshal(white)
	bJson, err2 := json.Marshal(black)
	if err1 != nil || err2 != nil {
		return "", "", false
	}
	return string(wJson), string(bJson), true
}
func (s *ChessService) createMatchFoundPayload(matchId string, white, black models.Player, fen string, maxMoveTime int64) ws.MatchFoundPayload {
	return ws.MatchFoundPayload{
		MatchId:     matchId,
		WhitePlayer: white,
		BlackPlayer: black,
		Fen:         fen,
		MaxMoveTime: maxMoveTime,
	}
}

func (s *ChessService) createGameStatePayload(matchId string, fen string, currentTurn string, moves []string, whiteTime, blackTime int64, maxMoveTime int64) ws.GameStatePayload {
	return ws.GameStatePayload{
		MatchId:       matchId,
		Fen:           fen,
		CurrentTurn:   currentTurn,
		MoveNotation:  moves,
		WhiteTimeLeft: whiteTime,
		BlackTimeLeft: blackTime,
		MaxMoveTime:   maxMoveTime,
	}
}

func (s *ChessService) createMatchCompletePayload(matchId, winnerId string, isDraw bool, reason enum.GameEndReason) ws.MatchCompletePayload {
	return ws.MatchCompletePayload{MatchId: matchId, WinnerId: winnerId, IsDraw: isDraw, Reason: reason}
}

func (s *ChessService) getMatchPlayers(matchId string) (models.Player, models.Player, bool) {
	playersMap, err := s.redisService.GetPlayers(matchId)
	if err != nil || len(playersMap) == 0 {
		return models.Player{}, models.Player{}, false
	}

	var white, black models.Player
	if json.Unmarshal([]byte(playersMap["WHITE"]), &white) != nil ||
		json.Unmarshal([]byte(playersMap["BLACK"]), &black) != nil {
		return models.Player{}, models.Player{}, false
	}

	return white, black, true
}

func (s *ChessService) cleanupMatchTimer(matchId string) {
	if val, ok := s.matchTimers.Load(matchId); ok {
		if matchTimer, matches := val.(*chessgame.MatchTimer); matches {
			matchTimer.Stop()
		}
		s.matchTimers.Delete(matchId)
	}
}

func (s *ChessService) updateAndGetTimers(matchId string) (int64, int64) {
	if val, ok := s.matchTimers.Load(matchId); ok {
		if matchTimer, matches := val.(*chessgame.MatchTimer); matches {
			matchTimer.SwitchTurn()
			return matchTimer.WhiteTimer.GetRemaining().Milliseconds(), matchTimer.BlackTimer.GetRemaining().Milliseconds()
		}
	}
	return 0, 0
}

func (s *ChessService) cleanupMatch(matchId string) {
	s.cleanupMatchTimer(matchId)

	_ = s.redisService.CleanData(matchId)
}

func (s *ChessService) determineGameReason(status chess.Method, isGameOver bool) (enum.GameEndReason, bool) {
	switch status {
	case chess.Checkmate:
		return enum.ReasonCheckmate, false
	case chess.Stalemate:
		return enum.ReasonStalemate, true
	case chess.InsufficientMaterial:
		return enum.ReasonInsufficient, true
	case chess.FivefoldRepetition, chess.ThreefoldRepetition, chess.FiftyMoveRule:
		return enum.ReasonDrawAgreement, true
	default:
		if isGameOver {
			return enum.ReasonDrawAgreement, true
		}
		return enum.GameEndReason(""), false
	}
}
