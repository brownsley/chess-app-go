package request

type UpdateRegionRequest struct {
	UserID  string `json:"user_id"`
	Country string `json:"country"`
}

type MatchEndUpdate struct {
	UserID    string
	IsWin     bool
	IsDraw    bool
	UpdateElo int
}
