package user

import (
	"github.com/tronget/human-touch/shared/storage"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	CreateUser(user *User) error
	GetUserByEmail(email string) (*User, error)
}

type repository struct {
	db *storage.DB
}

func NewRepository(db *storage.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateUser(u *User) error {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	res, err := r.db.Exec(`INSERT INTO users (email, password, name) VALUES ($1, $2, $3)`, u.Email, string(hashed), u.Name)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	_ = id
	return nil
}

func (r *repository) GetUserByEmail(email string) (*User, error) {
	var u User
	err := r.db.Get(&u, "SELECT id, email, password, name FROM users WHERE email=$1", email)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// func (r *repository) IsEmailTaken(email string) (bool, error) {
// 	var exists bool
// 	err := r.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", email)
// 	if err != nil {
// 		return false, err
// 	}
// 	return exists, nil
// }
