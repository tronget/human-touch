package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tronget/human-touch/shared/storage"
)

func JWT(next http.Handler) http.Handler {
	secret := []byte(os.Getenv("JWT_SECRET"))
	if len(secret) == 0 {
		log.Fatal("Missing `JWT_SECRET` in environment variables")
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// validate token if present and set X-User-ID header
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		authToken, ok := strings.CutPrefix(auth, "Bearer ")
		if !ok {
			http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(authToken, func(t *jwt.Token) (any, error) {
			return secret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid claims", http.StatusUnauthorized)
			return
		}

		if uid, ok := claims["user_id"].(float64); !ok {
			http.Error(w, "Invalid user id", http.StatusUnauthorized)
			return
		} else {
			r.Header.Set("X-User-ID", fmt.Sprintf("%d", int64(uid)))
		}
		next.ServeHTTP(w, r)
	})
}

func YandexToken(db *storage.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		client := &http.Client{Timeout: 5 * time.Second}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}

			var email string
			db.Get(&email, "SELECT email from app_user WHERE token=$1", auth)

			req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, "https://login.yandex.ru/info?format=json", nil)
			if err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}
			req.Header.Set("Authorization", auth)

			resp, err := client.Do(req)
			if err != nil {
				http.Error(w, "Failed to verify token", http.StatusUnauthorized)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			var info struct {
				DefaultEmail string `json:"default_email"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
				http.Error(w, "Failed to parse token info", http.StatusUnauthorized)
				return
			}

			if info.DefaultEmail == "" {
				http.Error(w, "Invalid info payload", http.StatusUnauthorized)
				return
			}

			r.Header.Set("X-Yandex-Email", info.DefaultEmail)

			next.ServeHTTP(w, r)
		})
	}
}

func IsExistingUser(db *storage.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			email := r.Header.Get("X-Yandex-Email")

			var exists bool

			row := db.QueryRow("SELECT EXISTS(SELECT 1 FROM app_user WHERE email=$1)", email)
			if err := row.Scan(&exists); err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				log.Println("serialize user existence check:", err)
				return
			}

			if !exists {
				http.Error(w, "User does not exist", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// func IsBanned(db *storage.DB) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			email := r.Header.Get("X-Yandex-Email")

// 			var bannedTill time.Time

// 			row := db.QueryRow(`SELECT banned_till FROM app_user WHERE email=$1`, email)
// 			if err := row.Scan(&bannedTill); err != nil {
// 				http.Error(w, "Internal error", http.StatusInternalServerError)
// 				log.Println("serialize banned check:", err)
// 				return
// 			}

// 			isBanned := !bannedTill.IsZero() && bannedTill.After(time.Now())

// 			if isBanned {
// 				http.Error(w, "User is banned", http.StatusForbidden)
// 				return
// 			}

// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }
