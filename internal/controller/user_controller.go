package controller

import (
	"encoding/json"
	"game-server/db"
	"game-server/internal/request"
	"game-server/internal/response"
	"net/http"

	"gorm.io/gorm"
)

type UserController struct {
	db *gorm.DB
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{db: db}
}

func (c *UserController) GetUserDetails(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "Missing user_id parameter", http.StatusBadRequest)
		return
	}
	var user db.User
	if err := c.db.First(&user, "user_id = ?", userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.UserResponse{
		ID:        user.ID,
		UserID:    user.UserID,
		GoogleID:  user.GoogleID,
		Email:     user.Email,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		Elo:       user.Elo,
		Country:   user.Country,
	})
}

func (c *UserController) ProcessMatchEnd(req request.MatchEndUpdate) error {
	updates := map[string]interface{}{
		"games_played": gorm.Expr("games_played + ?", 1),
		"elo":          gorm.Expr("elo + ?", req.UpdateElo),
	}
	if req.IsDraw {
		updates["draws"] = gorm.Expr("draws + ?", 1)
	} else if req.IsWin {
		updates["wins"] = gorm.Expr("wins + ?", 1)

	} else {
		updates["losses"] = gorm.Expr("losses + ?", 1)
	}

	result := c.db.Model(&db.User{}).Where("user_id = ?", req.UserID).Updates(updates)
	return result.Error
}

func (c *UserController) UpdateRegion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req request.UpdateRegionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Country == "" || req.UserID == "" {
		http.Error(w, "user_id and country are required", http.StatusBadRequest)
		return
	}

	result := c.db.Model(&db.User{}).Where("user_id = ?", req.UserID).Update("country", req.Country)
	if result.Error != nil {
		http.Error(w, "Failed to update country", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Region updated successfully",
		"country": req.Country,
	})
}
