package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type TokenValidator interface {
	ValidateToken(tokenStr string) (userID string, err error)
}

type MoveHandlerFunc func(move MovePayload)
type ResignHandlerFunc func(matchId string, resign ResignPayload)
type AcceptInviteHandlerFunc func(invite InvitePayload)
type MatchCompleteHandlerFunc func(roomID string, complete MatchCompletePayload)

type Room struct {
	ID      string
	Clients map[string]*websocket.Conn
	mu      sync.Mutex
}

type RoomManager struct {
	mu             sync.Mutex
	LobbyClients   map[string]*websocket.Conn
	MatchRooms     map[string]*Room
	Redis          *redis.Client
	TokenValidator TokenValidator

	OnMove          MoveHandlerFunc
	OnResign        ResignHandlerFunc
	OnAcceptInvite  AcceptInviteHandlerFunc
	OnMatchComplete MatchCompleteHandlerFunc
}

func NewRoomManager(redisClient *redis.Client, validator TokenValidator) *RoomManager {
	return &RoomManager{
		LobbyClients:   make(map[string]*websocket.Conn),
		MatchRooms:     make(map[string]*Room),
		Redis:          redisClient,
		TokenValidator: validator,
	}
}

func (rm *RoomManager) StartRedisSubscriber() {
	if rm.Redis == nil {
		return
	}
	ctx := context.Background()
	sub := rm.Redis.Subscribe(ctx, "global_match_channel")
	log.Println("Redis Subscriber started listening on 'global_match_channel'...")

	go func() {
		ch := sub.Channel()
		for msg := range ch {
			var wrapper struct {
				TargetIDs []string        `json:"targetIds"`
				Message   json.RawMessage `json:"message"`
			}
			if err := json.Unmarshal([]byte(msg.Payload), &wrapper); err == nil {
				for _, id := range wrapper.TargetIDs {
					rm.SendRawMessageToUser(id, wrapper.Message)
					rm.SendRawMessageToMatch(id, wrapper.Message)
				}
			}
		}
	}()
}
