package models

type UserResponse struct {
	ID        uint   `json:"id"`
	UserID    string `json:"user_id"`
	GoogleID  string `json:"google_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Elo       int    `json:"elo"`
	Country   string `json:"country"`
}

type AuthResponse struct {
	Message      string       `json:"message"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}
