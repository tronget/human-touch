package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	// wrapper around sqlx.DB
	X *sqlx.DB
}

func (db *DB) CreateUser(u *User) error {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	res, err := db.X.Exec(`INSERT INTO users (email, password, name) VALUES ($1, $2, $3)`, u.Email, string(hashed), u.Name)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId() // lib/pq doesn't support LastInsertId, you can use QueryRow RETURNING id in production
	_ = id
	return nil
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}
	// validate minimal
	if in.Email == "" || in.Password == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}
	// get DB from context
	db := r.Context().Value(CtxDBKey).(*DB)
	_, err := db.X.Exec(`INSERT INTO users (email, password, name) VALUES ($1, $2, $3)`, in.Email, hashPassword(in.Password), in.Name)
	if err != nil {
		http.Error(w, "db err: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func hashPassword(p string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(h)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}
	db := r.Context().Value(CtxDBKey).(*DB)
	var user User
	err := db.X.Get(&user, "SELECT id, email, password, name FROM users WHERE email=$1", in.Email)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)) != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	// make token
	secret := []byte(os.Getenv("JWT_SECRET"))
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokStr, _ := token.SignedString(secret)
	json.NewEncoder(w).Encode(map[string]string{"token": tokStr})
}
