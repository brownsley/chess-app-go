package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"game-server/db"
	"game-server/internal/response"
	"game-server/internal/service"

	"gorm.io/gorm"
)

type AuthHandler struct {
	db         *gorm.DB
	jwtService service.JWTService
	httpClient *http.Client
}

func NewAuthHandler(db *gorm.DB, jwtService service.JWTService) *AuthHandler {
	return &AuthHandler{
		db:         db,
		jwtService: jwtService,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (h *AuthHandler) HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IdToken string `json:"id_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	verifyURL := fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", url.QueryEscape(req.IdToken))

	res, err := h.httpClient.Get(verifyURL)
	if err != nil || res.StatusCode != http.StatusOK {
		http.Error(w, "Unauthorized: Invalid ID Token", http.StatusUnauthorized)
		return
	}
	defer res.Body.Close()

	var tokenInfo struct {
		Email   string `json:"email"`
		Name    string `json:"name"`
		UserId  string `json:"sub"`
		Picture string `json:"picture"`
		Aud     string `json:"aud"`
	}
	if err := json.NewDecoder(res.Body).Decode(&tokenInfo); err != nil || tokenInfo.Email == "" {
		http.Error(w, "Failed to parse token info", http.StatusInternalServerError)
		return
	}

	expectedClientID := os.Getenv("GOOGLE_CLIENT_ID")
	if expectedClientID != "" && tokenInfo.Aud != expectedClientID {
		http.Error(w, "Unauthorized: Audience mismatch", http.StatusUnauthorized)
		return
	}

	var user db.User
	isNewUser := false

	result := h.db.Where("google_id = ?", tokenInfo.UserId).First(&user)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			isNewUser = true

			user = db.User{
				GoogleID:  tokenInfo.UserId,
				Email:     tokenInfo.Email,
				Name:      tokenInfo.Name,
				AvatarURL: tokenInfo.Picture,
				Elo:       1200,
				Country:   "",
			}

			if createErr := h.db.Create(&user).Error; createErr != nil {
				log.Printf("[AUTH ERROR] Failed to create user for Google ID %s: %v", tokenInfo.UserId, createErr)
				http.Error(w, "Failed to create user account", http.StatusInternalServerError)
				return
			}

			log.Printf("[NEW USER] Registered successfully | Custom ID: %s | Google ID: %s", user.UserID, user.GoogleID)
		} else {
			log.Printf("[DATABASE ERROR] Failed to query user %s: %v", tokenInfo.UserId, result.Error)
			http.Error(w, "Database search error", http.StatusInternalServerError)
			return
		}
	} else {
		h.db.Model(&user).Select("Name", "AvatarURL").Updates(db.User{
			Name:      tokenInfo.Name,
			AvatarURL: tokenInfo.Picture,
		})

		user.Name = tokenInfo.Name
		user.AvatarURL = tokenInfo.Picture

		log.Printf("[EXISTING USER] Logged in | Custom ID: %s | DB ID: %d | Email: %s", user.UserID, user.ID, user.Email)
	}

	accessToken, refreshToken, err := h.jwtService.GenerateTokenPair(
		user.UserID,
		user.Email,
		user.Name,
	)
	if err != nil {
		log.Printf("[AUTH ERROR] Failed to generate token pair for User ID %s: %v", user.UserID, err)
		http.Error(w, "Failed to issue authentication tokens", http.StatusInternalServerError)
		return
	}

	response := response.AuthResponse{
		Message:      "Successfully authenticated",
		UserID:       user.UserID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IsNewUser:    isNewUser,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	claims, err := h.jwtService.ValidateToken(req.RefreshToken)
	if err != nil || claims.Subject != "refresh" {
		http.Error(w, "Unauthorized: Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	accessToken, newRefreshToken, err := h.jwtService.GenerateTokenPair(
		claims.UserID,
		claims.Email,
		claims.Name,
	)
	if err != nil {
		log.Printf("[AUTH ERROR] Token refresh failed for User ID %s: %v", claims.UserID, err)
		http.Error(w, "Failed to generate token pair", http.StatusInternalServerError)
		return
	}

	log.Printf("[TOKEN REFRESH] Token pair re-issued for User ID: %s", claims.UserID)

	response := map[string]interface{}{
		"user_id":       claims.UserID,
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
