package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tronget/communication-service/storage"
)

func CreateMessage(w http.ResponseWriter, r *http.Request) {
	var msg Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}

	db := r.Context().Value(CtxDBKey).(*storage.DB)
	uid := r.Context().Value(CtxUserID).(int64)
	now := time.Now()
	_, err := db.Exec(
		`INSERT INTO messages (conversation_id, sender_id, receiver_id, content, created_at) VALUES ($1,$2,$3,$4,$5)`,
		msg.ConversationID, uid, msg.ReceiverID, msg.Content, now,
	)
	if err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	BroadcastToUser(msg.ReceiverID, map[string]any{
		"type":       "new_message",
		"sender_id":  uid,
		"content":    msg.Content,
		"created_at": now,
	})
	w.WriteHeader(http.StatusCreated)
}

func GetMessages(w http.ResponseWriter, r *http.Request) {
	convID := chi.URLParam(r, "convId")
	// query messages for convID...
	_ = convID
	w.Write([]byte("[]"))
}
