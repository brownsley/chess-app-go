package routes

import (
	"game-server/internal/controller"
	"net/http"
)

func registerUserRoutes(mux *http.ServeMux, userHandler *controller.UserController) {
	mux.HandleFunc("/api/user/profile", userHandler.GetUserDetails)
	mux.HandleFunc("/api/user/region", userHandler.UpdateRegion)

}
