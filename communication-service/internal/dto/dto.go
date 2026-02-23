package dto

import (
	"database/sql"
	"time"
)

type MessageResponse struct {
	ID          int64     `json:"id"`
	ResponseID  int64     `json:"response_id"`
	SenderID    int64     `json:"sender_id"`
	ReceiverID  int64     `json:"receiver_id"`
	Text        string    `json:"text,omitempty"`
	ImageBase64 string    `json:"image_base64,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateMessageRequest struct {
	Text        string `json:"text"`
	ImageBase64 string `json:"image_base64"`
}

type ResponseChat struct {
	ResponseID        int64          `db:"response_id" json:"response_id"`
	ServiceID         int64          `db:"service_id" json:"service_id"`
	ServiceTitle      string         `db:"service_title" json:"service_title"`
	SenderID          int64          `db:"sender_id" json:"sender_id"`
	OwnerID           int64          `db:"owner_id" json:"owner_id"`
	ResponseCreatedAt time.Time      `db:"response_created_at" json:"response_created_at"`
	LastMessageID     sql.NullInt64  `db:"last_message_id" json:"last_message_id"`
	LastMessageAt     sql.NullTime   `db:"last_message_at" json:"last_message_at"`
	LastMessageText   sql.NullString `db:"last_message_text" json:"last_message_text"`
}
