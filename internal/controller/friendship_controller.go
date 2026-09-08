package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"game-server/db"
	"game-server/internal/response"

	"gorm.io/gorm"
)

type FriendshipController struct {
	db *gorm.DB
}

func NewFriendshipController(db *gorm.DB) *FriendshipController {
	return &FriendshipController{db: db}
}

type FriendRequestPayload struct {
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
}

func (c *FriendshipController) SendFriendRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	var payload FriendRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if payload.SenderID == payload.ReceiverID {
		http.Error(w, "cannot send friend request to yourself", http.StatusBadRequest)
		return
	}

	var count int64
	if err := c.db.Model(&db.User{}).Where("user_id = ?", payload.ReceiverID).Count(&count).Error; err != nil {
		http.Error(w, fmt.Sprintf("failed to check receiver user: %v", err), http.StatusInternalServerError)
		return
	}
	if count == 0 {
		http.Error(w, "receiver user not found", http.StatusNotFound)
		return
	}

	var existing db.Friendship
	err := c.db.Where(
		"(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		payload.SenderID, payload.ReceiverID, payload.ReceiverID, payload.SenderID,
	).First(&existing).Error

	if err == nil {
		http.Error(w, "friend request or friendship already exists", http.StatusBadRequest)
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		http.Error(w, fmt.Sprintf("failed to check existing friendship: %v", err), http.StatusInternalServerError)
		return
	}

	friendship := db.Friendship{
		UserID:   payload.SenderID,
		FriendID: payload.ReceiverID,
		Status:   db.StatusPending,
	}

	if err := c.db.Create(&friendship).Error; err != nil {
		http.Error(w, fmt.Sprintf("failed to send friend request: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Friend request sent successfully"})
}

func (c *FriendshipController) AcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	var payload FriendRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	err := c.db.Transaction(func(tx *gorm.DB) error {
		var friendship db.Friendship

		err := tx.Where("user_id = ? AND friend_id = ? AND status = ?",
			payload.SenderID, payload.ReceiverID, db.StatusPending).First(&friendship).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("friend request not found or already processed")
			}
			return fmt.Errorf("failed to find friend request: %w", err)
		}

		if err := tx.Model(&friendship).Update("status", db.StatusAccepted).Error; err != nil {
			return fmt.Errorf("failed to accept friend request: %w", err)
		}

		reverseFriendship := db.Friendship{
			UserID:   payload.ReceiverID,
			FriendID: payload.SenderID,
			Status:   db.StatusAccepted,
		}

		if err := tx.Where("user_id = ? AND friend_id = ?", payload.ReceiverID, payload.SenderID).
			FirstOrCreate(&reverseFriendship).Error; err != nil {
			return fmt.Errorf("failed to create reverse friendship record: %w", err)
		}
		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Friend request accepted"})
}

func (c *FriendshipController) RejectFriendRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	var payload FriendRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	err := c.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
			payload.SenderID, payload.ReceiverID, payload.ReceiverID, payload.SenderID).
			Delete(&db.Friendship{})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("friendship record not found")
		}

		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Friend request rejected or friend removed successfully"})
}

func (c *FriendshipController) GetFriendList(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id query parameter is required", http.StatusBadRequest)
		return
	}

	status := r.URL.Query().Get("status")

	var friends []response.FriendResponse

	if status == "pending" {
		err := c.db.Table("friendships").
			Select("users.user_id, users.name, users.avatar_url, users.elo, users.country, friendships.status").
			Joins("JOIN users ON users.user_id = friendships.user_id").
			Where("friendships.friend_id = ? AND friendships.status = ?", userID, db.StatusPending).
			Scan(&friends).Error

		if err != nil {
			http.Error(w, fmt.Sprintf("failed to fetch pending list: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		query := c.db.Table("friendships").
			Select("users.user_id, users.name, users.avatar_url, users.elo, users.country, friendships.status").
			Joins("JOIN users ON users.user_id = friendships.friend_id").
			Where("friendships.user_id = ?", userID)

		if status != "" && status != "ALL" {
			query = query.Where("friendships.status = ?", status)
		}

		if err := query.Scan(&friends).Error; err != nil {
			http.Error(w, fmt.Sprintf("failed to fetch friend list: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(friends)
}
