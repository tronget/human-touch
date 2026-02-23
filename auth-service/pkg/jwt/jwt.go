package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtToken string

type User interface {
	GetName() string
	GetID() int64
}

func Generate(u User, secret []byte) (JwtToken, error) {
	if len(secret) == 0 {
		return "", errors.New("missing `JWT_SECRET` in environment variables")
	}
	expire := time.Now().Add(time.Hour * 24)

	claims := jwt.MapClaims{
		"sub":     u.GetName(),
		"user_id": u.GetID(),
		"exp":     expire.Unix(),
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	s, err := token.SignedString(secret)
	if err != nil {
		return "", errors.New("could not sign token")
	}

	return JwtToken(s), nil
}
