package db

import "time"

type FriendshipStatus string

const (
	StatusPending  FriendshipStatus = "pending"
	StatusAccepted FriendshipStatus = "accepted"
	StatusBlocked  FriendshipStatus = "blocked"
)

type Friendship struct {
	ID        uint             `gorm:"primaryKey" json:"id"`
	UserID    string           `gorm:"type:varchar(12);not null;index:idx_user_friend,unique" json:"user_id"`
	FriendID  string           `gorm:"type:varchar(12);not null;index:idx_user_friend,unique" json:"friend_id"`
	Status    FriendshipStatus `gorm:"type:varchar(20);default:'pending';not null" json:"status"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`

	User   User `gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Friend User `gorm:"foreignKey:FriendID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}
