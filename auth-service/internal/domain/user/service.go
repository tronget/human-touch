package user

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/tronget/human-touch/auth-service/pkg/jwt"
	"github.com/tronget/human-touch/shared/storage"
	"golang.org/x/crypto/bcrypt"
)

type UserDto struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Service interface {
	RegisterUser(ctx context.Context, email, password, name string) (error, int)
	LoginUser(ctx context.Context, email, password string, jwtSecret []byte) (jwt.JwtToken, error, int)
	GetUserByID(ctx context.Context, id string) (*UserDto, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) RegisterUser(ctx context.Context, email, password, name string) (error, int) {
	if email == "" || password == "" || name == "" {
		slog.Warn("Invalid user data", "email", email, "name", name)
		return fmt.Errorf("Invalid user data"), http.StatusBadRequest
	}

	user := NewUser(email, password, name)

	if err := s.repo.CreateUser(ctx, user); err != nil {
		if storage.IsUniqueViolation(err) {
			slog.Warn("Email already taken during registration", "email", user.Email)
			return fmt.Errorf("Email already taken"), http.StatusConflict
		}

		slog.Error("Creating user in db", "error", err.Error())
		return fmt.Errorf("Error during user registration"), http.StatusInternalServerError
	}

	return nil, http.StatusCreated
}

func (s *service) LoginUser(ctx context.Context, email, password string, jwtSecret []byte) (jwt.JwtToken, error, int) {
	if email == "" || password == "" {
		slog.Warn("Invalid login data", "email", email)
		return "", fmt.Errorf("Invalid login data"), http.StatusBadRequest
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		slog.Warn("User not found by email during login",
			"email", email,
			"error", err.Error(),
		)
		return "", fmt.Errorf("Invalid credentials"), http.StatusUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		slog.Warn("Invalid password attempt",
			"email", email,
			"error", err.Error(),
		)
		return "", fmt.Errorf("Invalid credentials"), http.StatusUnauthorized
	}

	token, err := jwt.Generate(user, jwtSecret)
	if err != nil || token == "" {
		slog.Error("Could not generate token", "error", err.Error())
		return "", fmt.Errorf("could not generate token: %s", err.Error()), http.StatusInternalServerError
	}

	return token, nil, http.StatusOK
}

func (s *service) GetUserByID(ctx context.Context, id string) (*UserDto, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		slog.Warn("User not found by ID", "id", id, "error", err.Error())
		return nil, fmt.Errorf("User not found")
	}

	userDto := &UserDto{
		ID:    id,
		Email: user.Email,
		Name:  user.Name,
	}
	return userDto, nil
}
