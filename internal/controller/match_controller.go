package controller

import (
	"encoding/json"
	"net/http"

	"game-server/internal/models"
	service "game-server/internal/service/matching"
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
		http.Error(w, "Invalid request payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.PlayerId == "" || req.PlayerName == "" {
		http.Error(w, "playerId and playerName are required", http.StatusBadRequest)
		return
	}

	c.matchingService.JoinQueue(req.GameMode, req.Minutes, req.PlayerId, req.PlayerName, req.PlayerElo)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Joined queue successfully",
	})
}

func (c *MatchController) LeaveQueueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.LeaveQueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.PlayerId == "" || req.PlayerName == "" {
		http.Error(w, "playerId and playerName are required", http.StatusBadRequest)
		return
	}

	err := c.matchingService.LeaveFromMatchingQueue(req.GameMode, req.PlayerId, req.PlayerName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Left queue successfully",
	})
}
