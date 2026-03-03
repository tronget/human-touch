package dialogue

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tronget/human-touch/communication-service/internal/userctx"
)

func GetActiveDialoguesHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := userctx.UserID(r.Context())
		if !ok || uid == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userID, _ := strconv.ParseInt(uid, 10, 64)

		dialogues, err := service.GetActiveDialogues(userID)
		if err != nil {
			slog.Error("failed to get active dialogues", "user_id", userID, "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(dialogues)
		if err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func CloseDialogueHandler(service Service) http.HandlerFunc {
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

		if err, code := service.CloseDialogue(dialogueID, userID); err != nil {
			http.Error(w, err.Error(), code)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"closed"}`))
	}
}
