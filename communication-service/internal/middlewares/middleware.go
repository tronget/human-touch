package middlewares

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tronget/human-touch/communication-service/internal/constants"
	"github.com/tronget/human-touch/communication-service/internal/models"
	"github.com/tronget/human-touch/shared/storage"
)

func WithDB(db *storage.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), constants.CtxDBKey, db)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func WithUser(db *storage.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := resolveUser(r, db)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if user.BannedTill.Valid && user.BannedTill.Time.After(time.Now()) {
				http.Error(w, "user is banned", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), constants.CtxUserKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func resolveUser(r *http.Request, db *storage.DB) (models.User, error) {
	var user models.User

	if uidHeader := strings.TrimSpace(r.Header.Get("X-User-ID")); uidHeader != "" {
		uid, err := strconv.ParseInt(uidHeader, 10, 64)
		if err != nil {
			return user, errors.New("invalid user id")
		}
		err = db.Get(&user, `SELECT id, email, role, banned_till FROM app_user WHERE id=$1`, uid)
		if errors.Is(err, sql.ErrNoRows) {
			return user, errors.New("user with id " + uidHeader + " not found")
		}
		return user, err
	}

	email := strings.TrimSpace(r.Header.Get("X-Yandex-Email"))
	if email == "" {
		return user, errors.New("missing user identity")
	}

	err := db.Get(&user, `SELECT id, email, role, banned_till FROM app_user WHERE email=$1`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return user, errors.New("user not found")
	}
	return user, err
}
