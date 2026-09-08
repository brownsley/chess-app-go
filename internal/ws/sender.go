package ws

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (rm *RoomManager) HandleLobbyWebSocket(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "Bad Request: Missing userId", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Lobby Upgrade Error:", err)
		return
	}

	rm.mu.Lock()
	rm.LobbyClients[userID] = conn
	rm.mu.Unlock()

	log.Printf("Lobby Connected: Player %s\n", userID)

	defer func() {
		conn.Close()
		rm.mu.Lock()
		delete(rm.LobbyClients, userID)
		rm.mu.Unlock()
		log.Printf("Lobby Disconnected: Player %s\n", userID)
	}()

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		switch msg.Type {
		case TypeInvite:
			payloadBytes, _ := json.Marshal(msg.Payload)
			var invite InvitePayload
			if err := json.Unmarshal(payloadBytes, &invite); err == nil {
				rm.SendInvite(invite)
			}
		case TypeAcceptInvite:
			payloadBytes, _ := json.Marshal(msg.Payload)
			var invite InvitePayload
			if err := json.Unmarshal(payloadBytes, &invite); err == nil && rm.OnAcceptInvite != nil {
				rm.OnAcceptInvite(invite)
			}
		}
	}
}

func (rm *RoomManager) HandleMatchWebSocket(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "Bad Request: Missing userId", http.StatusBadRequest)
		return
	}

	matchID := r.PathValue("matchId")
	if matchID == "" {
		http.Error(w, "Match ID required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Match Upgrade Error:", err)
		return
	}

	rm.mu.Lock()
	room, exists := rm.MatchRooms[matchID]
	if !exists {
		room = &Room{
			ID:      matchID,
			Clients: make(map[string]*websocket.Conn),
		}
		rm.MatchRooms[matchID] = room
	}
	room.mu.Lock()
	room.Clients[userID] = conn
	room.mu.Unlock()
	rm.mu.Unlock()

	log.Printf("Match Connected: User %s in Match %s\n", userID, matchID)

	defer func() {
		conn.Close()
		rm.removeMatchClient(matchID, userID)
		log.Printf("Match Disconnected: User %s from Match %s\n", userID, matchID)
	}()

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		switch msg.Type {
		case TypeMove:
			payloadBytes, err := json.Marshal(msg.Payload)
			if err != nil {
				continue
			}
			var move MovePayload
			if err := json.Unmarshal(payloadBytes, &move); err != nil {
				continue
			}
			if rm.OnMove != nil {
				rm.OnMove(move)
			}

		case TypeResign:
			payloadBytes, _ := json.Marshal(msg.Payload)
			var resign ResignPayload
			if err := json.Unmarshal(payloadBytes, &resign); err == nil && rm.OnResign != nil {
				rm.OnResign(matchID, resign)
			}

		case TypeOfferDraw:
			payloadBytes, _ := json.Marshal(msg.Payload)
			var draw OfferDrawPayload
			if err := json.Unmarshal(payloadBytes, &draw); err == nil && rm.OnOfferDraw != nil {
				rm.OnOfferDraw(matchID, draw)
			}
		case TypeAcceptDraw:
			payloadBytes, _ := json.Marshal(msg.Payload)
			var accept AcceptDrawPayload
			if err := json.Unmarshal(payloadBytes, &accept); err == nil && rm.OnAcceptDraw != nil {
				rm.OnAcceptDraw(matchID, accept)
			}

		case TypeDeclineDraw:
			payloadBytes, _ := json.Marshal(msg.Payload)
			var decline DeclineDrawPayload
			if err := json.Unmarshal(payloadBytes, &decline); err == nil && rm.OnDeclineDraw != nil {
				rm.OnDeclineDraw(matchID, decline)
			}
		case TypeMatchComplete:
			payloadBytes, _ := json.Marshal(msg.Payload)
			var complete MatchCompletePayload
			if err := json.Unmarshal(payloadBytes, &complete); err == nil && rm.OnMatchComplete != nil {
				rm.OnMatchComplete(matchID, complete)
			}
		}
	}
}

func (rm *RoomManager) removeMatchClient(matchID string, userID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.MatchRooms[matchID]
	if !exists {
		return
	}

	room.mu.Lock()
	delete(room.Clients, userID)
	isEmpty := len(room.Clients) == 0
	room.mu.Unlock()

	if isEmpty {
		delete(rm.MatchRooms, matchID)
	}
}
