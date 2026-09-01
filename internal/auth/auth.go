package auth

import (
	"encoding/json"
	"fmt"
	"game-server/internal/service"
	"net/http"
	"net/url"
)

type AuthHandler struct {
	jwtService service.JWTService
}

func NewAuthHandler(jwtService service.JWTService) *AuthHandler {
	return &AuthHandler{jwtService: jwtService}
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
	res, err := http.Get(verifyURL)
	if err != nil || res.StatusCode != http.StatusOK {
		http.Error(w, "Unauthorized: Invalid ID Token", http.StatusUnauthorized)
		return
	}
	defer res.Body.Close()

	var tokenInfo struct {
		Email  string `json:"email"`
		Name   string `json:"name"`
		UserId string `json:"sub"`
	}
	if err := json.NewDecoder(res.Body).Decode(&tokenInfo); err != nil || tokenInfo.Email == "" {
		http.Error(w, "Failed to parse token info", http.StatusInternalServerError)
		return
	}

	accessToken, refreshToken, err := h.jwtService.GenerateTokenPair(
		tokenInfo.UserId,
		tokenInfo.Email,
		tokenInfo.Name,
	)
	if err != nil {
		http.Error(w, "Failed to issue authentication tokens", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message":       "Successfully authenticated",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"email":         tokenInfo.Email,
		"name":          tokenInfo.Name,
		"google_id":     tokenInfo.UserId,
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
		http.Error(w, "Failed to generate token pair", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
