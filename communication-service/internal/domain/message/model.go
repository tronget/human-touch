package message

import "time"

type Message struct {
	ID         int64     `db:"id" json:"id"`
	DialogueID int64     `db:"dialogue_id" json:"dialogue_id"`
	SenderID   int64     `db:"sender_id" json:"sender_id"`
	Content    string    `db:"content" json:"content"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}
