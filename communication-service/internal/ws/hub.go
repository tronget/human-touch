package ws

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/tronget/human-touch/communication-service/internal/domain/dialogue"
	"github.com/tronget/human-touch/communication-service/internal/domain/message"
	"github.com/tronget/human-touch/communication-service/internal/matchmaking"
	"github.com/tronget/human-touch/shared/jwtx"
)

var allowedOrigins = map[string]bool{
	"http://localhost:3000": true,
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return allowedOrigins[origin]
	},
}

type InMessageEvent string
type OutMessageEvent string

const (
	SearchEvent        InMessageEvent = "search"
	CancelSearchEvent  InMessageEvent = "cancel_search"
	MessageEvent       InMessageEvent = "message"
	CloseDialogueEvent InMessageEvent = "close_dialogue"
)

const (
	MatchedEvent         OutMessageEvent = "matched"
	NewMessageEvent      OutMessageEvent = "new_message"
	DialogueClosedEvent  OutMessageEvent = "dialogue_closed"
	ErrorEvent           OutMessageEvent = "error"
	SearchCancelledEvent OutMessageEvent = "search_cancelled"
)

type IncomingMessage struct {
	Type InMessageEvent  `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

type OutgoingMessage struct {
	Type OutMessageEvent `json:"type"`
	Data any             `json:"data,omitempty"`
}

type SendMessageData struct {
	DialogueID int64  `json:"dialogue_id"`
	Content    string `json:"content"`
}

type CloseDialogueData struct {
	DialogueID int64 `json:"dialogue_id"`
}

type Hub interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type hub struct {
	mu          sync.RWMutex
	connections map[int64]*websocket.Conn

	jwtSecret  []byte
	queue      matchmaking.Queue
	msgService message.Service
	dlgService dialogue.Service
}

func NewHub(jwtSecret []byte, queue matchmaking.Queue, msgService message.Service, dlgService dialogue.Service) Hub {
	return &hub{
		connections: make(map[int64]*websocket.Conn),
		jwtSecret:   jwtSecret,
		queue:       queue,
		msgService:  msgService,
		dlgService:  dlgService,
	}
}

func (h *hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, err := h.authenticateWS(r)
	if err != nil {
		slog.Warn("ws auth failed", "err", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "user_id", userID, "err", err)
		return
	}

	h.mu.Lock()
	if old, exists := h.connections[userID]; exists {
		old.Close()
	}
	h.connections[userID] = conn
	h.mu.Unlock()

	slog.Info("websocket connected", "user_id", userID)

	defer func() {
		h.mu.Lock()
		delete(h.connections, userID)
		h.mu.Unlock()
		conn.Close()
		h.queue.Cancel(userID)
		slog.Info("websocket disconnected", "user_id", userID)
	}()

	h.readLoop(conn, userID)
}

func (h *hub) readLoop(conn *websocket.Conn, userID int64) {
	for {
		var msg IncomingMessage
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				slog.Warn("websocket read error", "user_id", userID, "err", err)
			}
			return
		}

		switch msg.Type {
		case SearchEvent:
			h.handleSearch(conn, userID)
		case CancelSearchEvent:
			h.handleCancelSearch(conn, userID)
		case MessageEvent:
			h.handleMessage(conn, userID, msg.Data)
		case CloseDialogueEvent:
			h.handleCloseDialogue(conn, userID, msg.Data)
		default:
			h.sendJSON(conn, OutgoingMessage{
				Type: ErrorEvent,
				Data: "unknown message type",
			})
		}
	}
}

func (h *hub) handleSearch(conn *websocket.Conn, userID int64) {
	resultCh := h.queue.Enqueue(userID)

	go func() {
		result, ok := <-resultCh
		if !ok {
			return
		}

		matchData := map[string]any{
			"dialogue_id": result.Dialogue.ID,
			"partner_id":  result.PartnerID,
		}

		h.mu.RLock()
		if c, exists := h.connections[userID]; exists {
			h.sendJSON(c, OutgoingMessage{Type: MatchedEvent, Data: matchData})
		}
		h.mu.RUnlock()
	}()
}

func (h *hub) handleCancelSearch(conn *websocket.Conn, userID int64) {
	h.queue.Cancel(userID)
	h.sendJSON(conn, OutgoingMessage{Type: SearchCancelledEvent})
}

func (h *hub) handleMessage(conn *websocket.Conn, userID int64, data json.RawMessage) {
	var req SendMessageData
	if err := json.Unmarshal(data, &req); err != nil {
		h.sendJSON(conn, OutgoingMessage{Type: ErrorEvent, Data: "invalid message data"})
		return
	}

	msg, svcErr, _ := h.msgService.SendMessage(req.DialogueID, userID, req.Content)
	if svcErr != nil {
		h.sendJSON(conn, OutgoingMessage{Type: ErrorEvent, Data: svcErr.Error()})
		return
	}

	outgoing := OutgoingMessage{Type: NewMessageEvent, Data: msg}

	d, err := h.dlgService.GetDialogue(req.DialogueID)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, uid := range []int64{d.User1ID, d.User2ID} {
		if c, exists := h.connections[uid]; exists {
			h.sendJSON(c, outgoing)
		}
	}
}

func (h *hub) handleCloseDialogue(conn *websocket.Conn, userID int64, data json.RawMessage) {
	var req CloseDialogueData
	if err := json.Unmarshal(data, &req); err != nil {
		h.sendJSON(conn, OutgoingMessage{Type: ErrorEvent, Data: "invalid data"})
		return
	}

	d, err := h.dlgService.GetDialogue(req.DialogueID)
	if err != nil {
		h.sendJSON(conn, OutgoingMessage{Type: ErrorEvent, Data: "dialogue not found"})
		return
	}

	if svcErr, _ := h.dlgService.CloseDialogue(req.DialogueID, userID); svcErr != nil {
		h.sendJSON(conn, OutgoingMessage{Type: ErrorEvent, Data: svcErr.Error()})
		return
	}

	closedData := map[string]any{
		"dialogue_id": req.DialogueID,
		"closed_by":   userID,
	}
	outgoing := OutgoingMessage{Type: DialogueClosedEvent, Data: closedData}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, uid := range []int64{d.User1ID, d.User2ID} {
		if c, exists := h.connections[uid]; exists {
			h.sendJSON(c, outgoing)
		}
	}
}

func (h *hub) sendJSON(conn *websocket.Conn, msg OutgoingMessage) {
	if err := conn.WriteJSON(msg); err != nil {
		slog.Warn("failed to send websocket message", "err", err)
	}
}

func (h *hub) authenticateWS(r *http.Request) (int64, error) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		return 0, fmt.Errorf("missing token query param")
	}

	uid, err := jwtx.ValidateAndExtractUserID(tokenStr, h.jwtSecret)
	if err != nil {
		return 0, err
	}
	return uid, nil
}
