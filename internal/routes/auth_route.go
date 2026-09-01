package routes

import (
	"net/http"

	"game-server/internal/auth"
)

func registerAuthRoutes(mux *http.ServeMux, authHandler *auth.AuthHandler) {
	mux.HandleFunc("POST /api/auth/google", authHandler.HandleGoogleLogin)
	mux.HandleFunc("OPTIONS /api/auth/google", authHandler.HandleGoogleLogin)

	mux.HandleFunc("POST /api/auth/refresh", authHandler.HandleRefreshToken)
	mux.HandleFunc("OPTIONS /api/auth/refresh", authHandler.HandleRefreshToken)
}
