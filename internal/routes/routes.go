package routes

import (
	"net/http"

	"game-server/internal/auth"
	"game-server/internal/controller"
	"game-server/internal/ws"
)

func RegisterRoutes(
	mux *http.ServeMux,
	authHandler *auth.AuthHandler,
	matchController *controller.MatchController,
	roomManager *ws.RoomManager,
) {
	registerAuthRoutes(mux, authHandler)
	registerMatchRoutes(mux, matchController)

	mux.HandleFunc("GET /ws/lobby", roomManager.HandleLobbyWebSocket)

	mux.HandleFunc("GET /ws/match/{matchId}", roomManager.HandleMatchWebSocket)

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Server is running on: " + r.Host))
	})
}
