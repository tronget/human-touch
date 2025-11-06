package main

import "time"

type Message struct {
	ID             int64     `db:"id" json:"id"`
	ConversationID int64     `db:"conversation_id" json:"conversation_id"`
	FromUserID     int64     `db:"from_user_id" json:"from_user_id"`
	ToUserID       int64     `db:"to_user_id" json:"to_user_id"`
	Text           string    `db:"text" json:"text"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}
