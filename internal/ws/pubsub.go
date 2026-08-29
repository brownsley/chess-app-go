package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type MoveHandlerFunc func(move MovePayload)
type ResignHandlerFunc func(matchId string, resign ResignPayload)
type AcceptInviteHandlerFunc func(invite InvitePayload)
type MatchCompleteHandlerFunc func(roomID string, complete MatchCompletePayload)

type Room struct {
	ID      string
	Clients []*websocket.Conn
	mu      sync.Mutex
}

type RoomManager struct {
	mu              sync.Mutex
	OnMove          MoveHandlerFunc
	OnResign        ResignHandlerFunc
	OnAcceptInvite  AcceptInviteHandlerFunc
	OnMatchComplete MatchCompleteHandlerFunc

	Rooms map[string]*Room
	Redis *redis.Client
}

func NewRoomManager(redisClient *redis.Client) *RoomManager {
	return &RoomManager{
		Rooms: make(map[string]*Room),
		Redis: redisClient,
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
					rm.SendRawMessageToRoom(id, wrapper.Message)
				}
			}
		}
	}()
}

func (rm *RoomManager) BroadcastToRedis(targetIDs []string, msgType MessageType, payload interface{}) {
	for _, id := range targetIDs {
		rm.SendMessageByRoomID(id, msgType, payload)
	}

	if rm.Redis == nil {
		return
	}

	msg := Message{
		Type:    msgType,
		Payload: payload,
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return
	}

	wrapper := map[string]interface{}{
		"targetIds": targetIDs,
		"message":   json.RawMessage(msgBytes),
	}
	wrapperBytes, err := json.Marshal(wrapper)
	if err != nil {
		return
	}

	_ = rm.Redis.Publish(context.Background(), "global_match_channel", wrapperBytes).Err()
}
