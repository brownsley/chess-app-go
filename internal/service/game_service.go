package service

import (
	"log"
	"math"

	"game-server/internal/request"
)

type UserStatsRepository interface {
	ProcessMatchEnd(req request.MatchEndUpdate) error
}

type GameService struct {
	userRepo UserStatsRepository
}

func NewGameService(userRepo UserStatsRepository) *GameService {
	return &GameService{
		userRepo: userRepo,
	}
}

func (s *GameService) ProcessGameResult(player1ID string, p1Elo int, player2ID string, p2Elo int, winnerID string, isDraw bool) error {
	var scoreA float64

	if isDraw {
		scoreA = 0.5
	} else if winnerID == player1ID {
		scoreA = 1.0
	} else {
		scoreA = 0.0
	}

	_, _, changeA, changeB := calculateElo(p1Elo, p2Elo, scoreA)

	err1 := s.userRepo.ProcessMatchEnd(request.MatchEndUpdate{
		UserID:    player1ID,
		IsWin:     !isDraw && winnerID == player1ID,
		IsDraw:    isDraw,
		UpdateElo: changeA,
	})

	err2 := s.userRepo.ProcessMatchEnd(request.MatchEndUpdate{
		UserID:    player2ID,
		IsWin:     !isDraw && winnerID == player2ID,
		IsDraw:    isDraw,
		UpdateElo: changeB,
	})

	if err1 != nil {
		log.Printf("[GAME ERROR] Failed to update P1 (%s): %v", player1ID, err1)
		return err1
	}
	if err2 != nil {
		log.Printf("[GAME ERROR] Failed to update P2 (%s): %v", player2ID, err2)
		return err2
	}

	log.Printf("[GAME END] P1 (%s): Change %d | P2 (%s): Change %d", player1ID, changeA, player2ID, changeB)
	return nil
}

func calculateElo(ratingA, ratingB int, scoreA float64) (newRatingA, newRatingB, changeA, changeB int) {
	const kFactor = 32.0

	expectedA := 1.0 / (1.0 + math.Pow(10, float64(ratingB-ratingA)/400.0))
	expectedB := 1.0 - expectedA

	scoreB := 1.0 - scoreA

	deltaA := int(math.Round(kFactor * (scoreA - expectedA)))
	deltaB := int(math.Round(kFactor * (scoreB - expectedB)))

	return ratingA + deltaA, ratingB + deltaB, deltaA, deltaB
}
