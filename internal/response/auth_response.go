package response

type UserResponse struct {
	ID          uint   `json:"id"`
	UserID      string `json:"user_id"`
	GoogleID    string `json:"google_id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	AvatarURL   string `json:"avatar_url"`
	Elo         int    `json:"elo"`
	Country     string `json:"country"`
	GamesPlayed int    `json:"games_played"`
	Wins        int    `json:"wins"`
	Losses      int    `json:"losses"`
	Draws       int    `json:"draws"`
}

type AuthResponse struct {
	Message      string `json:"message"`
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IsNewUser    bool   `json:"is_new_user"`
}
