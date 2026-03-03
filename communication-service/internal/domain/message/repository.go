package message

import (
	"github.com/tronget/human-touch/shared/storage"
)

type Repository interface {
	Create(dialogueID, senderID int64, content string) (*Message, error)
	GetByDialogueID(dialogueID int64, limit, offset int) ([]Message, error)
}

type repository struct {
	db *storage.DB
}

func NewRepository(db *storage.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(dialogueID, senderID int64, content string) (*Message, error) {
	var m Message
	err := r.db.Get(&m,
		`INSERT INTO messages (dialogue_id, sender_id, content)
		 VALUES ($1, $2, $3)
		 RETURNING id, dialogue_id, sender_id, content, created_at`,
		dialogueID, senderID, content,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *repository) GetByDialogueID(dialogueID int64, limit, offset int) ([]Message, error) {
	var messages []Message
	err := r.db.Select(&messages,
		`SELECT * FROM messages
		 WHERE dialogue_id=$1
		 ORDER BY created_at ASC
		 LIMIT $2 OFFSET $3`,
		dialogueID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	return messages, nil
}
