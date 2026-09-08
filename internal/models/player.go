package models

type PlayerSide string

const (
	White PlayerSide = "W"
	Black PlayerSide = "B"
)

type Player struct {
	ID      string     `json:"id"`
	Name    string     `json:"name"`
	Side    PlayerSide `json:"side"`
	Elo     int        `json:"elo"`
	Avatar  string     `json:"avatar"`
	Country string     `json:"country"`
}
