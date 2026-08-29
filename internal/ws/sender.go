package ws

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

func (rm *RoomManager) SendRawMessageToRoom(id string, message json.RawMessage) {
	rm.mu.Lock()
	room, exists := rm.Rooms[id]
	rm.mu.Unlock()

	if !exists {
		return
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	for _, client := range room.Clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("SendRawMessage error:", err)
		}
	}
}

func (rm *RoomManager) SendMessageByRoomID(roomID string, msgType MessageType, payload interface{}) {
	rm.mu.Lock()
	room, exists := rm.Rooms[roomID]
	rm.mu.Unlock()

	if !exists {
		return
	}
	room.mu.Lock()
	defer room.mu.Unlock()

	msg := Message{
		Type:    msgType,
		Payload: payload,
	}

	for _, client := range room.Clients {
		_ = client.WriteJSON(msg)
	}
}

func (rm *RoomManager) SendMatchFoundToBoth(matchId string, playerId1, playerId2 string, matchData MatchFoundPayload) {
	targetIDs := []string{playerId1, playerId2, matchId}
	rm.BroadcastToRedis(targetIDs, TypeMatchFound, matchData)
}

func (rm *RoomManager) SendInvite(invite InvitePayload) {
	targetID := []string{invite.OtherPlayer}
	rm.BroadcastToRedis(targetID, TypeInvite, invite)
}

func (rm *RoomManager) SendMatchMoveProcess(matchId string, targetIDs []string, gameState GameStatePayload) {
	rm.BroadcastToRedis([]string{matchId}, TypeMove, gameState)
}

func (rm *RoomManager) SendMatchComplete(matchId string, completePayload MatchCompletePayload) {
	rm.BroadcastToRedis([]string{matchId}, TypeMatchComplete, completePayload)
}
func (rm *RoomManager) SendResignProcess(matchId string, targetIDs []string, resignPayload ResignPayload) {
	rm.BroadcastToRedis([]string{matchId}, TypeResign, resignPayload)
}
