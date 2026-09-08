package db

import (
	"game-server/utils"
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UserID    string `gorm:"type:varchar(12);uniqueIndex;not null" json:"user_id"`
	GoogleID  string `gorm:"type:varchar(255);uniqueIndex;not null" json:"google_id"`
	Email     string `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Name      string `gorm:"type:varchar(255)" json:"name"`
	AvatarURL string `gorm:"type:text" json:"avatar_url"`

	Elo         int    `gorm:"default:1200;not null;index" json:"elo"`
	Country     string `gorm:"type:varchar(3);default:'MM'" json:"country"`
	GamesPlayed int    `gorm:"default:0;not null" json:"games_played"`
	Wins        int    `gorm:"default:0;not null" json:"wins"`
	Losses      int    `gorm:"default:0;not null" json:"losses"`
	Draws       int    `gorm:"default:0;not null" json:"draws"`

	Friends []Friendship `gorm:"-" json:"friends,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.UserID == "" {
		u.UserID = utils.UserIdGenerate()
	}
	return
}
