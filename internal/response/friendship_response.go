package response

type FriendResponse struct {
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Elo       int    `json:"elo"`
	Country   string `json:"country"`
	Status    string `json:"status"`
}
