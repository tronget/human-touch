package user

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/tronget/human-touch/auth-service/pkg/jwt"
	"github.com/tronget/human-touch/shared/storage"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	RegisterUser(email, password, name string) (error, int)
	LoginUser(email, password string, jwtSecret []byte) (jwt.JwtToken, error, int)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) RegisterUser(email, password, name string) (error, int) {
	if email == "" || password == "" || name == "" {
		slog.Warn("Invalid user data", "email", email, "name", name)
		return fmt.Errorf("Invalid user data"), http.StatusBadRequest
	}

	user := NewUser(email, password, name)

	if err := s.repo.CreateUser(user); err != nil {
		if storage.IsUniqueViolation(err) {
			slog.Warn("Email already taken during registration", "email", user.Email)
			return fmt.Errorf("Email already taken"), http.StatusConflict
		}

		slog.Error("Creating user in db", "error", err.Error())
		return fmt.Errorf("Error during user registration"), http.StatusInternalServerError
	}

	return nil, http.StatusCreated
}

func (s *service) LoginUser(email, password string, jwtSecret []byte) (jwt.JwtToken, error, int) {
	if email == "" || password == "" {
		slog.Warn("Invalid login data", "email", email)
		return "", fmt.Errorf("Invalid login data"), http.StatusBadRequest
	}

	user, err := s.repo.GetUserByEmail(email)
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
