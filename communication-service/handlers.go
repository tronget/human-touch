package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	// wrapper around sqlx.DB
	X *sqlx.DB
}

func CreateMessage(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ConversationID int64  `json:"conversation_id"`
		ToUserID       int64  `json:"to_user_id"`
		Text           string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}

	db := r.Context().Value(CtxDBKey).(*DB)
	uid := r.Context().Value(CtxUserID).(int64)
	now := time.Now()

	_, err := db.X.Exec(
		`INSERT INTO messages (conversation_id, from_user_id, to_user_id, text, created_at) VALUES ($1,$2,$3,$4,$5)`,
		in.ConversationID, uid, in.ToUserID, in.Text, now,
	)
	if err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	BroadcastToUser(in.ToUserID, map[string]any{
		"type": "new_message",
		"from": uid,
		"text": in.Text,
		"time": now,
	})
	w.WriteHeader(http.StatusCreated)
}

func GetMessages(w http.ResponseWriter, r *http.Request) {
	convID := chi.URLParam(r, "convId")
	// query messages for convID...
	_ = convID
	w.Write([]byte("[]"))
}
