package dialogue

import (
	"fmt"
	"log/slog"
	"net/http"
)

type Service interface {
	CreateDialogue(user1ID, user2ID int64) (*Dialogue, error)
	GetDialogue(id int64) (*Dialogue, error)
	GetActiveDialogues(userID int64) ([]Dialogue, error)
	CloseDialogue(dialogueID int64, requestingUserID int64) (error, int)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateDialogue(user1ID, user2ID int64) (*Dialogue, error) {
	d, err := s.repo.Create(user1ID, user2ID)
	if err != nil {
		slog.Error("failed to create dialogue", "user1", user1ID, "user2", user2ID, "err", err)
		return nil, err
	}
	slog.Info("dialogue created", "dialogue_id", d.ID, "user1", user1ID, "user2", user2ID)
	return d, nil
}

func (s *service) GetDialogue(id int64) (*Dialogue, error) {
	return s.repo.GetByID(id)
}

func (s *service) GetActiveDialogues(userID int64) ([]Dialogue, error) {
	return s.repo.GetActiveByUserID(userID)
}

func (s *service) CloseDialogue(dialogueID int64, requestingUserID int64) (error, int) {
	d, err := s.repo.GetByID(dialogueID)
	if err != nil {
		slog.Warn("dialogue not found for close", "dialogue_id", dialogueID, "err", err)
		return fmt.Errorf("dialogue not found"), http.StatusNotFound
	}

	if d.User1ID != requestingUserID && d.User2ID != requestingUserID {
		slog.Warn("user tried to close foreign dialogue",
			"dialogue_id", dialogueID, "user_id", requestingUserID)
		return fmt.Errorf("forbidden"), http.StatusForbidden
	}

	if !d.IsActive {
		return fmt.Errorf("dialogue already closed"), http.StatusBadRequest
	}

	if err := s.repo.Close(dialogueID); err != nil {
		slog.Error("failed to close dialogue", "dialogue_id", dialogueID, "err", err)
		return fmt.Errorf("internal error"), http.StatusInternalServerError
	}

	slog.Info("dialogue closed", "dialogue_id", dialogueID, "by_user", requestingUserID)
	return nil, http.StatusOK
}
