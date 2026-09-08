package ws

import (
	"game-server/internal/enum"
	"game-server/internal/game"
	"game-server/internal/models"
)

type MessageType string

type ColorChoice string

const (
	ChoiceWhite  ColorChoice = "WHITE"
	ChoiceBlack  ColorChoice = "BLACK"
	ChoiceRandom ColorChoice = "RANDOM"
)

const (
	TypeMatchFound    MessageType = "match_found"
	TypeInvite        MessageType = "invite"
	TypeMove          MessageType = "move"
	TypeAcceptInvite  MessageType = "accept_invite"
	TypeResign        MessageType = "resign"
	TypeOfferDraw     MessageType = "offer_draw"
	TypeAcceptDraw    MessageType = "accept_draw"
	TypeDeclineDraw   MessageType = "decline_draw"
	TypeMatchComplete MessageType = "match_complete"
)

type IncomingMessage struct {
	Type    string      `json:"type"`
	Payload MovePayload `json:"payload"`
}

type InvitePayload struct {
	SenderID        string         `json:"senderId"`
	SenderName      string         `json:"senderName"`
	SenderCountry   string         `json:"senderCountry"`
	SenderAvatar    string         `json:"senderAvatar"`
	SenderElo       int            `json:"senderElo"`
	ReceiverID      string         `json:"receiverId"`
	MatchType       game.MatchType `json:"matchType"`
	ColorPreference ColorChoice    `json:"colorPreference"`
}

type MovePayload struct {
	MatchId   string `json:"matchId"`
	From      string `json:"from"`
	To        string `json:"to"`
	Promotion string `json:"promotion"`
	PlayerID  string `json:"playerId"`
}

type GameStatePayload struct {
	MatchId       string   `json:"matchId"`
	Fen           string   `json:"fen"`
	CurrentTurn   string   `json:"currentTurn"`
	MoveNotation  []string `json:"moveNotation"`
	MaxMoveTime   int64    `json:"maxMoveTime"`
	WhiteTimeLeft int64    `json:"white_time_left"`
	BlackTimeLeft int64    `json:"black_time_left"`
}

type ResignPayload struct {
	PlayerID string `json:"playerId"`
}

type OfferDrawPayload struct {
	PlayerID string `json:"playerId"`
}

type AcceptDrawPayload struct {
	PlayerID string `json:"playerId"`
}

type DeclineDrawPayload struct {
	PlayerID string `json:"playerId"`
}

type MatchFoundPayload struct {
	MatchId     string        `json:"matchId"`
	WhitePlayer models.Player `json:"whitePlayer"`
	BlackPlayer models.Player `json:"blackPlayer"`
	Fen         string        `json:"fen"`
	MaxMoveTime int64         `json:"maxMoveTime"`
}

type MatchCompletePayload struct {
	MatchId  string             `json:"matchId"`
	WinnerId string             `json:"winnerId,omitempty"`
	IsDraw   bool               `json:"isDraw"`
	Reason   enum.GameEndReason `json:"reason"`
}
type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}
