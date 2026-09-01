package routes

import (
	"net/http"

	"game-server/internal/controller"
)

func registerMatchRoutes(mux *http.ServeMux, matchController *controller.MatchController) {
	mux.HandleFunc("POST /api/match/join", matchController.JoinQueueHandler)
	mux.HandleFunc("POST /api/match/leave", matchController.LeaveQueueHandler)
	mux.HandleFunc("OPTIONS /api/match/join", matchController.JoinQueueHandler)
	mux.HandleFunc("OPTIONS /api/match/leave", matchController.LeaveQueueHandler)
}
