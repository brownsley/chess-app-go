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

func (rm *RoomManager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade Error:", err)
		return
	}
	defer conn.Close()

	roomID := r.PathValue("roomId")
	if roomID == "" {
		log.Println("Room ID is empty")
		return
	}

	rm.mu.Lock()
	room, exist := rm.Rooms[roomID]
	if !exist {
		room = &Room{
			ID:      roomID,
			Clients: []*websocket.Conn{},
		}
		rm.Rooms[roomID] = room
	}
	rm.mu.Unlock()

	room.mu.Lock()
	room.Clients = append(room.Clients, conn)
	room.mu.Unlock()

	log.Printf("Client connected to Room/Player: %s\n", roomID)

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Client disconnected from Room/Player: %s\n", roomID)
			rm.removeClient(roomID, conn)
			break
		}

		switch msg.Type {
		case TypeMove:
			payloadBytes, _ := json.Marshal(msg.Payload)
			var move MovePayload
			if err := json.Unmarshal(payloadBytes, &move); err == nil {
				if rm.OnMove != nil {
					rm.OnMove(move)
				}
			}
		case TypeResign:
			payloadBytes, _ := json.Marshal(msg.Payload)
			var resign ResignPayload
			if err := json.Unmarshal(payloadBytes, &resign); err == nil {
				if rm.OnResign != nil {
					rm.OnResign(roomID, resign)
				}
			}

		case TypeInvite:
			payloadBytes, _ := json.Marshal(msg.Payload)
			var invite InvitePayload
			if err := json.Unmarshal(payloadBytes, &invite); err == nil {
				rm.SendInvite(invite)
				log.Printf("Received Invite from %s to %s\n", invite.ChallengerID, invite.OtherPlayer)
			}
		case TypeAcceptInvite:
			payloadBytes, _ := json.Marshal(msg.Payload)
			var invite InvitePayload
			if err := json.Unmarshal(payloadBytes, &invite); err == nil {
				if rm.OnAcceptInvite != nil {
					rm.OnAcceptInvite(invite)
				}
			}

		case TypeMatchComplete:
			payloadBytes, _ := json.Marshal(msg.Payload)
			var complete MatchCompletePayload
			if err := json.Unmarshal(payloadBytes, &complete); err == nil {
				if rm.OnMatchComplete != nil {
					rm.OnMatchComplete(roomID, complete)
				}
			}
		default:
			log.Printf("Unknown message type: %s\n", msg.Type)
		}
	}
}

func (rm *RoomManager) removeClient(roomID string, conn *websocket.Conn) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	room, exist := rm.Rooms[roomID]
	if !exist {
		return
	}
	room.mu.Lock()
	defer room.mu.Unlock()

	filtered := []*websocket.Conn{}
	for _, c := range room.Clients {
		if c != conn {
			filtered = append(filtered, c)
		}
	}
	room.Clients = filtered
	if len(room.Clients) == 0 {
		delete(rm.Rooms, roomID)
	}
}
