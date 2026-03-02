package user

import (
	"context"

	"github.com/tronget/human-touch/shared/storage"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
}

type repository struct {
	db *storage.DB
}

func NewRepository(db *storage.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateUser(ctx context.Context, u *User) error {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	res, err := r.db.ExecContext(ctx, `INSERT INTO users (email, password, name) VALUES ($1, $2, $3)`, u.Email, string(hashed), u.Name)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	_ = id
	return nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.db.GetContext(ctx, &u, "SELECT id, email, password, name FROM users WHERE email=$1", email)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) GetUserByID(ctx context.Context, id string) (*User, error) {
	var u User
	err := r.db.GetContext(ctx, &u, "SELECT id, email, password, name FROM users WHERE id=$1", id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
