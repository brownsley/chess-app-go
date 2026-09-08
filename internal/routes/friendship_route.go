package routes

import (
	"game-server/internal/controller"
	"net/http"
)

func registerFriendshipRoutes(mux *http.ServeMux, friendshipController *controller.FriendshipController) {
	mux.HandleFunc("POST /api/friends/request", friendshipController.SendFriendRequest)
	mux.HandleFunc("OPTIONS /api/friends/request", friendshipController.SendFriendRequest)

	mux.HandleFunc("POST /api/friends/accept", friendshipController.AcceptFriendRequest)
	mux.HandleFunc("OPTIONS /api/friends/accept", friendshipController.AcceptFriendRequest)

	mux.HandleFunc("POST /api/friends/reject", friendshipController.RejectFriendRequest)
	mux.HandleFunc("OPTIONS /api/friends/reject", friendshipController.RejectFriendRequest)

	mux.HandleFunc("GET /api/friends", friendshipController.GetFriendList)
	mux.HandleFunc("OPTIONS /api/friends", friendshipController.GetFriendList)
}
