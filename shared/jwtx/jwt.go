package jwtx

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func ValidateToken(tokenStr string, jwtSecret []byte) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return token, nil
}

func ValidateClaims(token *jwt.Token) (jwt.MapClaims, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	if exp, ok := claims["exp"].(float64); !ok || exp < float64(time.Now().Unix()) {
		return nil, fmt.Errorf("token has expired")
	}

	uid, ok := claims["user_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing user_id in claims")
	}
	if uid <= 0 {
		return nil, fmt.Errorf("invalid user_id in claims")
	}

	return claims, nil
}

func ExtractUserID(claims jwt.MapClaims) (int64, error) {
	uid, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("missing user_id in claims")
	}
	if uid <= 0 {
		return 0, fmt.Errorf("invalid user_id in claims")
	}

	return int64(uid), nil
}

func ValidateAndExtractUserID(tokenStr string, jwtSecret []byte) (int64, error) {
	tokenStr, ok := strings.CutPrefix(tokenStr, "Bearer ")
	if !ok {
		return 0, fmt.Errorf("invalid token format")
	}

	token, err := ValidateToken(tokenStr, jwtSecret)
	if err != nil {
		return 0, err
	}

	claims, err := ValidateClaims(token)
	if err != nil {
		return 0, fmt.Errorf("invalid claims: %w", err)
	}

	uid, err := ExtractUserID(claims)
	if err != nil {
		return 0, err
	}

	return uid, nil
}
