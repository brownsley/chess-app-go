package ws

import (
	"context"
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

func (rm *RoomManager) SendRawMessageToUser(userID string, message json.RawMessage) {
	rm.mu.Lock()
	conn, exists := rm.LobbyClients[userID]
	rm.mu.Unlock()

	if exists {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Println("SendRawMessageToUser error:", err)
		}
	}
}

func (rm *RoomManager) SendRawMessageToMatch(matchID string, message json.RawMessage) {
	rm.mu.Lock()
	room, exists := rm.MatchRooms[matchID]
	rm.mu.Unlock()

	if !exists {
		return
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	for _, client := range room.Clients {
		_ = client.WriteMessage(websocket.TextMessage, message)
	}
}

func (rm *RoomManager) SendMessageByUserID(userID string, msgType MessageType, payload interface{}) {
	rm.mu.Lock()
	conn, exists := rm.LobbyClients[userID]
	rm.mu.Unlock()

	if exists {
		_ = conn.WriteJSON(Message{Type: msgType, Payload: payload})
	}
}

func (rm *RoomManager) SendMessageByMatchID(matchID string, msgType MessageType, payload interface{}) {
	rm.mu.Lock()
	room, exists := rm.MatchRooms[matchID]
	rm.mu.Unlock()

	if !exists {
		return
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	msg := Message{Type: msgType, Payload: payload}
	for _, client := range room.Clients {
		_ = client.WriteJSON(msg)
	}
}

func (rm *RoomManager) BroadcastToRedis(targetIDs []string, msgType MessageType, payload interface{}) {
	if rm.Redis == nil {
		for _, id := range targetIDs {
			rm.SendMessageByUserID(id, msgType, payload)
			rm.SendMessageByMatchID(id, msgType, payload)
		}
		return
	}

	msg := Message{Type: msgType, Payload: payload}
	msgBytes, _ := json.Marshal(msg)

	wrapper := map[string]interface{}{
		"targetIds": targetIDs,
		"message":   json.RawMessage(msgBytes),
	}
	wrapperBytes, _ := json.Marshal(wrapper)

	_ = rm.Redis.Publish(context.Background(), "global_match_channel", wrapperBytes).Err()
}

func (rm *RoomManager) SendMatchFoundToBoth(matchId string, playerId1, playerId2 string, matchData MatchFoundPayload) {
	targetIDs := []string{playerId1, playerId2}
	rm.BroadcastToRedis(targetIDs, TypeMatchFound, matchData)
}

func (rm *RoomManager) SendInvite(invite InvitePayload) {
	targetIDs := []string{invite.ReceiverID}
	rm.BroadcastToRedis(targetIDs, TypeInvite, invite)
}

func (rm *RoomManager) SendMatchMoveProcess(matchId string, gameState GameStatePayload) {
	rm.BroadcastToRedis([]string{matchId}, TypeMove, gameState)
}

func (rm *RoomManager) SendResignProcess(matchId string, resignPayload ResignPayload) {
	rm.BroadcastToRedis([]string{matchId}, TypeResign, resignPayload)
}

func (rm *RoomManager) SendMatchComplete(matchId string, completePayload MatchCompletePayload) {
	rm.BroadcastToRedis([]string{matchId}, TypeMatchComplete, completePayload)
}

func (rm *RoomManager) SendOfferDraw(matchId string, offerDraw OfferDrawPayload) {
	rm.BroadcastToRedis([]string{matchId}, TypeOfferDraw, offerDraw)
}

func (rm *RoomManager) SendAcceptDraw(matchId string, acceptDraw AcceptDrawPayload) {
	rm.BroadcastToRedis([]string{matchId}, TypeAcceptDraw, acceptDraw)
}

func (rm *RoomManager) SendDeclineDraw(matchId string, declineDraw DeclineDrawPayload) {
	rm.BroadcastToRedis([]string{matchId}, TypeDeclineDraw, declineDraw)
}
