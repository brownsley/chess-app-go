package http

import (
	"encoding/json"
	"game-server/internal/models"
	"game-server/internal/service"
	"net/http"
)

type MatchController struct {
	matchingService *service.MatchingService
}

func NewMatchController(ms *service.MatchingService) *MatchController {
	return &MatchController{
		matchingService: ms,
	}
}

func (c *MatchController) JoinQueueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.JoinQueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	c.matchingService.JoinQueue(req.GameMode, req.Minutes, req.PlayerId, req.PlayerElo)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Joined queue successfully"}`))
}

func (c *MatchController) LeaveQueueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.LeaveQueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := c.matchingService.LeaveFromMatchingQueue(req.GameMode, req.PlayerId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Left queue successfully"}`))
}
