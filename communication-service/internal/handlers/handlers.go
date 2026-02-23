package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/tronget/human-touch/communication-service/internal/constants"
	"github.com/tronget/human-touch/communication-service/internal/dto"
	"github.com/tronget/human-touch/communication-service/internal/models"
	"github.com/tronget/human-touch/communication-service/internal/ws"
	"github.com/tronget/human-touch/shared/storage"
)

type convoParticipants struct {
	ResponseID int64 `db:"id"`
	SenderID   int64 `db:"sender_id"`
	OwnerID    int64 `db:"owner_id"`
}

func GetChatsWhereUserIsSender(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value(constants.CtxDBKey).(*storage.DB)
	user := r.Context().Value(constants.CtxUserKey).(models.User)

	items, err := listChatsByRole(db, "sender", user.ID)
	if err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func GetChatsWhereUserIsOwner(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value(constants.CtxDBKey).(*storage.DB)
	user := r.Context().Value(constants.CtxUserKey).(models.User)

	items, err := listChatsByRole(db, "owner", user.ID)
	if err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func listChatsByRole(db *storage.DB, role string, userID int64) ([]dto.ResponseChat, error) {
	condition := ""
	switch role {
	case "sender":
		condition = "r.sender_id = $1"
	case "owner":
		condition = "s.owner_id = $1"
	default:
		return nil, fmt.Errorf("unknown role filter: %s", role)
	}

	query := fmt.Sprintf(`
SELECT
  r.id AS response_id,
  r.service_id,
  s.title AS service_title,
  r.sender_id,
  s.owner_id,
  r.created_at AS response_created_at,
  lm.id AS last_message_id,
  lm.created_at AS last_message_at,
  lm.message_text AS last_message_text
FROM response r
JOIN service s ON s.id = r.service_id
LEFT JOIN LATERAL (
    SELECT id, created_at, message_text
    FROM message
    WHERE response_id = r.id
    ORDER BY id DESC
    LIMIT 1
) lm ON TRUE
WHERE %s
ORDER BY COALESCE(lm.created_at, r.created_at) DESC`, condition)

	var items []dto.ResponseChat
	if err := db.Select(&items, query, userID); err != nil {
		return nil, err
	}

	return items, nil
}

func CreateMessage(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	text := strings.TrimSpace(req.Text)
	imgBase64 := strings.TrimSpace(req.ImageBase64)
	if text == "" && imgBase64 == "" {
		http.Error(w, "text or image is required", http.StatusBadRequest)
		return
	}
	if len(text) > 5000 {
		http.Error(w, "text exceeds 5000 characters", http.StatusBadRequest)
		return
	}

	var imageBytes []byte
	if imgBase64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(imgBase64)
		if err != nil {
			http.Error(w, "image must be base64", http.StatusBadRequest)
			return
		}
		if len(decoded) > 10*1024*1024 {
			http.Error(w, "image too large (max 10MB)", http.StatusRequestEntityTooLarge)
			return
		}
		imageBytes = decoded
	}

	responseID, err := strconv.ParseInt(chi.URLParam(r, "responseId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid response id", http.StatusBadRequest)
		return
	}

	db := r.Context().Value(constants.CtxDBKey).(*storage.DB)
	user := r.Context().Value(constants.CtxUserKey).(models.User)

	conv, err := loadConversation(db, responseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "conversation not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to load conversation", http.StatusInternalServerError)
		return
	}

	if user.ID != conv.SenderID && user.ID != conv.OwnerID {
		http.Error(w, "not a conversation participant", http.StatusForbidden)
		return
	}

	receiverID := conv.OwnerID
	if user.ID == conv.OwnerID {
		receiverID = conv.SenderID
	}

	var msg models.Message
	insert := `INSERT INTO message (response_id, sender_id, receiver_id, message_text, message_image)
			VALUES ($1,$2,$3,$4,$5)
			RETURNING id, response_id, sender_id, receiver_id, message_text, message_image, created_at`

	textVal := sql.NullString{String: text, Valid: text != ""}
	err = db.QueryRow(insert, responseID, user.ID, receiverID, textVal, imageBytes).
		Scan(&msg.ID, &msg.ResponseID, &msg.SenderID, &msg.ReceiverID, &msg.Text, &msg.Image, &msg.CreatedAt)
	if err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := toMessageResponse(msg)
	if receiverID != 0 {
		go ws.BroadcastToUser(receiverID, map[string]any{
			"type":    "new_message",
			"payload": resp,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func GetMessages(w http.ResponseWriter, r *http.Request) {
	responseID, err := strconv.ParseInt(chi.URLParam(r, "responseId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid response id", http.StatusBadRequest)
		return
	}

	db := r.Context().Value(constants.CtxDBKey).(*storage.DB)
	user := r.Context().Value(constants.CtxUserKey).(models.User)

	conv, err := loadConversation(db, responseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "conversation not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to load conversation", http.StatusInternalServerError)
		return
	}

	if user.ID != conv.SenderID && user.ID != conv.OwnerID {
		http.Error(w, "not a conversation participant", http.StatusForbidden)
		return
	}

	limit := 50
	if q := strings.TrimSpace(r.URL.Query().Get("limit")); q != "" {
		if v, err := strconv.Atoi(q); err == nil && v > 0 {
			if v > 200 {
				v = 200
			}
			limit = v
		}
	}

	var afterID int64
	if q := strings.TrimSpace(r.URL.Query().Get("after_id")); q != "" {
		if v, err := strconv.ParseInt(q, 10, 64); err == nil && v > 0 {
			afterID = v
		}
	}

	query := `SELECT id, response_id, sender_id, receiver_id, message_text, message_image, created_at FROM message WHERE response_id=$1`
	args := []any{responseID}
	argPos := 2
	if afterID > 0 {
		query += fmt.Sprintf(" AND id > $%d", argPos)
		args = append(args, afterID)
		argPos++
	}
	query += fmt.Sprintf(" ORDER BY id ASC LIMIT $%d", argPos)
	args = append(args, limit)

	var msgs []models.Message
	if err := db.Select(&msgs, query, args...); err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := make([]dto.MessageResponse, 0, len(msgs))
	for _, m := range msgs {
		resp = append(resp, toMessageResponse(m))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func loadConversation(db *storage.DB, responseID int64) (convoParticipants, error) {
	var conv convoParticipants
	query := `SELECT r.id, r.sender_id, s.owner_id
	FROM response r
	JOIN service s ON s.id = r.service_id
	WHERE r.id=$1`
	return conv, db.Get(&conv, query, responseID)
}

func toMessageResponse(m models.Message) dto.MessageResponse {
	dto := dto.MessageResponse{
		ID:         m.ID,
		ResponseID: m.ResponseID,
		SenderID:   m.SenderID,
		ReceiverID: m.ReceiverID,
		CreatedAt:  m.CreatedAt,
	}

	if m.Text.Valid {
		dto.Text = m.Text.String
	}
	if len(m.Image) > 0 {
		dto.ImageBase64 = base64.StdEncoding.EncodeToString(m.Image)
	}
	return dto
}
