package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	GoogleID     string  `gorm:"type:varchar(255);uniqueIndex;not null" json:"google_id"`
	Email        string  `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Username     string  `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	Name         string  `gorm:"type:varchar(255)" json:"name"`
	AvatarURL    string  `gorm:"type:text" json:"avatar_url"`
	AuthProvider string  `gorm:"type:varchar(50);default:'google'" json:"auth_provider"`
	Password     *string `gorm:"type:varchar(255)" json:"password"`

	Elo         int    `gorm:"default:1200;not null" json:"elo"`
	Country     string `gorm:"type:varchar(3);default:'MM'" json:"country"`
	GamesPlayed int    `gorm:"default:0;not null" json:"games_played"`
	Wins        int    `gorm:"default:0;not null" json:"wins"`
	Losses      int    `gorm:"default:0;not null" json:"losses"`
	Draws       int    `gorm:"default:0;not null" json:"draws"`
}
