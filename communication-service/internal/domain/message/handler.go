package message

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tronget/human-touch/communication-service/internal/userctx"
)

type SendMessageRequest struct {
	Content string `json:"content"`
}

func SendMessageHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := userctx.UserID(r.Context())
		if !ok || uid == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		userID, _ := strconv.ParseInt(uid, 10, 64)

		dialogueIDStr := chi.URLParam(r, "dialogueID")
		dialogueID, err := strconv.ParseInt(dialogueIDStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid dialogue id", http.StatusBadRequest)
			return
		}

		var req SendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request body", http.StatusBadRequest)
			slog.Error("bad send message request", "err", err)
			return
		}

		msg, svcErr, code := service.SendMessage(dialogueID, userID, req.Content)
		if svcErr != nil {
			http.Error(w, svcErr.Error(), code)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(msg)
	}
}

func GetMessagesHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := userctx.UserID(r.Context())
		if !ok || uid == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		userID, _ := strconv.ParseInt(uid, 10, 64)

		dialogueIDStr := chi.URLParam(r, "dialogueID")
		dialogueID, err := strconv.ParseInt(dialogueIDStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid dialogue id", http.StatusBadRequest)
			return
		}

		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil {
			limit = 50
		}
		offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
		if err != nil {
			offset = 0
		}

		messages, svcErr, code := service.GetMessages(dialogueID, userID, limit, offset)
		if svcErr != nil {
			http.Error(w, svcErr.Error(), code)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
	}
}
