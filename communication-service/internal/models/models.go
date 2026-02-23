package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID         int64        `db:"id"`
	Email      string       `db:"email"`
	Role       string       `db:"role"`
	BannedTill sql.NullTime `db:"banned_till"`
}

type Message struct {
	ID         int64          `db:"id" json:"id"`
	ResponseID int64          `db:"response_id" json:"response_id"`
	SenderID   int64          `db:"sender_id" json:"sender_id"`
	ReceiverID int64          `db:"receiver_id" json:"receiver_id"`
	Text       sql.NullString `db:"message_text" json:"-"`
	Image      []byte         `db:"message_image" json:"-"`
	CreatedAt  time.Time      `db:"created_at" json:"created_at"`
}
