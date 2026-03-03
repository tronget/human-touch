package matchmaking

import (
	"log/slog"
	"sync"

	"github.com/tronget/human-touch/communication-service/internal/domain/dialogue"
)

type MatchResult struct {
	Dialogue  *dialogue.Dialogue
	PartnerID int64
}

type waiter struct {
	UserID int64
	Result chan MatchResult
}

type Queue interface {
	Enqueue(userID int64) <-chan MatchResult
	Cancel(userID int64)
}

type queue struct {
	mu          sync.Mutex
	waiting     *waiter
	dialogueSvc dialogue.Service
}

func NewQueue(dialogueSvc dialogue.Service) Queue {
	return &queue{dialogueSvc: dialogueSvc}
}

func (q *queue) Enqueue(userID int64) <-chan MatchResult {
	ch := make(chan MatchResult, 1)
	w := &waiter{UserID: userID, Result: ch}

	q.mu.Lock()
	defer q.mu.Unlock()

	if q.waiting != nil && q.waiting.UserID == userID {
		return q.waiting.Result
	}

	if q.waiting != nil {
		other := q.waiting
		q.waiting = nil
		go q.createMatch(other, w)
		return ch
	}

	q.waiting = w
	slog.Info("user enqueued for matchmaking", "user_id", userID)
	return ch
}

func (q *queue) Cancel(userID int64) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.waiting != nil && q.waiting.UserID == userID {
		close(q.waiting.Result)
		q.waiting = nil
		slog.Info("user cancelled matchmaking", "user_id", userID)
	}
}

func (q *queue) createMatch(w1 *waiter, w2 *waiter) {
	d, err := q.dialogueSvc.CreateDialogue(w1.UserID, w2.UserID)
	if err != nil {
		slog.Error("failed to create dialogue during matchmaking",
			"user1", w1.UserID, "user2", w2.UserID, "err", err)
		close(w1.Result)
		close(w2.Result)
		return
	}

	slog.Info("match found", "dialogue_id", d.ID, "user1", w1.UserID, "user2", w2.UserID)

	w1.Result <- MatchResult{Dialogue: d, PartnerID: w2.UserID}
	w2.Result <- MatchResult{Dialogue: d, PartnerID: w1.UserID}
}
