package message

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/tronget/human-touch/communication-service/internal/domain/dialogue"
)

type Service interface {
	SendMessage(dialogueID, senderID int64, content string) (*Message, error, int)
	GetMessages(dialogueID int64, userID int64, limit, offset int) ([]Message, error, int)
}

type service struct {
	msgRepo         Repository
	dialogueService dialogue.Service
}

func NewService(msgRepo Repository, dialogueService dialogue.Service) Service {
	return &service{
		msgRepo:         msgRepo,
		dialogueService: dialogueService,
	}
}

func (s *service) SendMessage(dialogueID, senderID int64, content string) (*Message, error, int) {
	if content == "" {
		return nil, fmt.Errorf("message content is empty"), http.StatusBadRequest
	}

	d, err := s.dialogueService.GetDialogue(dialogueID)
	if err != nil {
		slog.Warn("dialogue not found for message", "dialogue_id", dialogueID, "err", err)
		return nil, fmt.Errorf("dialogue not found"), http.StatusNotFound
	}

	if !d.IsActive {
		return nil, fmt.Errorf("dialogue is closed"), http.StatusBadRequest
	}

	if d.User1ID != senderID && d.User2ID != senderID {
		slog.Warn("user tried to send message to foreign dialogue",
			"dialogue_id", dialogueID,
			"sender_id", senderID,
		)
		return nil, fmt.Errorf("forbidden"), http.StatusForbidden
	}

	msg, err := s.msgRepo.Create(dialogueID, senderID, content)
	if err != nil {
		slog.Error("failed to create message",
			"dialogue_id", dialogueID,
			"sender_id", senderID,
			"err", err,
		)
		return nil, fmt.Errorf("internal error"), http.StatusInternalServerError
	}

	slog.Debug("message sent",
		"message_id", msg.ID,
		"dialogue_id", dialogueID,
		"sender_id", senderID,
	)
	return msg, nil, http.StatusCreated
}

func (s *service) GetMessages(dialogueID int64, userID int64, limit, offset int) ([]Message, error, int) {
	d, err := s.dialogueService.GetDialogue(dialogueID)
	if err != nil {
		return nil, fmt.Errorf("dialogue not found"), http.StatusNotFound
	}

	if d.User1ID != userID && d.User2ID != userID {
		return nil, fmt.Errorf("forbidden"), http.StatusForbidden
	}

	if limit <= 0 || limit > 100000 {
		limit = 50
	}

	messages, err := s.msgRepo.GetByDialogueID(dialogueID, limit, offset)
	if err != nil {
		slog.Error("failed to get messages", "dialogue_id", dialogueID, "err", err)
		return nil, fmt.Errorf("internal error"), http.StatusInternalServerError
	}

	return messages, nil, http.StatusOK
}
