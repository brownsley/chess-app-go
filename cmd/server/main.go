package main

import (
	"crypto/tls"
	"flag"
	"log"
	nethttp "net/http"
	"os"
	"time"

	"game-server/db"
	"game-server/internal/auth"
	"game-server/internal/controller"
	"game-server/internal/routes"
	service "game-server/internal/service/auth"
	chess "game-server/internal/service/chess"
	redisService "game-server/internal/service/redis"

	match "game-server/internal/service/matching"
	"game-server/internal/ws"

	"github.com/redis/go-redis/v9"
)

func main() {

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "closing-grackle-177023.upstash.io:6379",
		Username: "default",
		Password: "gQAAAAAAArN_AAIgcDFlODQ1ZTQwMDFiMmQ0MGE3OTFlNzQwOWI0ZjdkNTQ2Zg",
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	})
	dsn := "host=ep-billowing-lab-b3geagfe-pooler.c-4.ap-southeast-1.aws.neon.tech port=5432 user=neondb_owner password=npg_4bmQRICs6wXg dbname=neondb sslmode=require"
	// dsn := "host=localhost port=5432 user=postgres password=secret dbname=postgres sslmode=disable"

	db, err := db.InitDB(dsn)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	redisService := redisService.NewRedisService(redisClient)
	roomManager := ws.NewRoomManager(redisClient, nil)
	go roomManager.StartRedisSubscriber()

	chessService := chess.NewChessService(redisService, roomManager)
	matchingService := match.NewMatchingService(redisService, chessService, roomManager)
	matchController := controller.NewMatchController(matchingService)

	jwtService := service.NewJWTService("your_secret_key_here", 15*time.Minute, 7*24*time.Hour)
	authHandler := auth.NewAuthHandler(db, jwtService)

	roomManager.OnMove = func(move ws.MovePayload) {
		chessService.ProcessMove(move)
	}

	roomManager.OnResign = func(roomID string, resign ws.ResignPayload) {
		chessService.ProcessResign(roomID, resign)
	}

	roomManager.OnAcceptInvite = func(invite ws.InvitePayload) {
		matchingService.StartFriendMatch(invite, true)
	}

	mux := nethttp.NewServeMux()
	routes.RegisterRoutes(mux, authHandler, matchController, roomManager)

	portFlag := flag.String("port", "", "Port to run server on")
	flag.Parse()

	port := *portFlag
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}

	log.Printf("Server started on :%s", port)
	if err := nethttp.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal("Server Error: ", err)
	}
}
