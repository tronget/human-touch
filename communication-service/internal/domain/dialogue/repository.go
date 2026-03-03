package dialogue

import (
	"github.com/tronget/human-touch/shared/storage"
)

type Repository interface {
	Create(user1ID, user2ID int64) (*Dialogue, error)
	GetByID(id int64) (*Dialogue, error)
	GetActiveByUserID(userID int64) ([]Dialogue, error)
	Close(id int64) error
}

type repository struct {
	db *storage.DB
}

func NewRepository(db *storage.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(user1ID, user2ID int64) (*Dialogue, error) {
	var d Dialogue
	err := r.db.Get(&d,
		`INSERT INTO dialogues (user1_id, user2_id)
		 VALUES ($1, $2)
		 RETURNING id, user1_id, user2_id, is_active, created_at, closed_at`,
		user1ID, user2ID,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *repository) GetByID(id int64) (*Dialogue, error) {
	var d Dialogue
	err := r.db.Get(&d, `SELECT * FROM dialogues WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *repository) GetActiveByUserID(userID int64) ([]Dialogue, error) {
	var dialogues []Dialogue
	err := r.db.Select(&dialogues,
		`SELECT * FROM dialogues
		 WHERE (user1_id=$1 OR user2_id=$1) AND is_active=true
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	return dialogues, nil
}

func (r *repository) Close(id int64) error {
	_, err := r.db.Exec(
		`UPDATE dialogues SET is_active=false, closed_at=now() WHERE id=$1`,
		id,
	)
	return err
}
