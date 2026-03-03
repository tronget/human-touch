package server

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tronget/human-touch/communication-service/internal/config"
	"github.com/tronget/human-touch/communication-service/internal/domain/dialogue"
	"github.com/tronget/human-touch/communication-service/internal/domain/message"
	"github.com/tronget/human-touch/communication-service/internal/matchmaking"
	"github.com/tronget/human-touch/communication-service/internal/middleware"
	"github.com/tronget/human-touch/communication-service/internal/ws"
	"github.com/tronget/human-touch/shared/storage"
)

type Server interface {
	Run() error
}

type server struct {
	cfg *config.Config
	db  *storage.DB
}

func NewServer(cfg *config.Config, db *storage.DB) Server {
	return &server{
		cfg: cfg,
		db:  db,
	}
}

func (s *server) Run() error {
	dialogueRepo := dialogue.NewRepository(s.db)
	messageRepo := message.NewRepository(s.db)

	dialogueService := dialogue.NewService(dialogueRepo)
	messageService := message.NewService(messageRepo, dialogueService)

	queue := matchmaking.NewQueue(dialogueService)

	hub := ws.NewHub([]byte(s.cfg.JwtSecret), queue, messageService, dialogueService)

	r := chi.NewRouter()
	r.Use(middleware.ExtractHeadersMiddleware)

	r.Get("/dialogues", dialogue.GetActiveDialoguesHandler(dialogueService))
	r.Post("/dialogues/{dialogueID}/close", dialogue.CloseDialogueHandler(dialogueService))
	r.Get("/dialogues/{dialogueID}/messages", message.GetMessagesHandler(messageService))
	r.Post("/dialogues/{dialogueID}/messages", message.SendMessageHandler(messageService))

	r.Get("/ws", hub.ServeHTTP)

	slog.Info("communication service listening on " + s.cfg.ServiceAddress)
	return http.ListenAndServe(s.cfg.ServiceAddress, r)
}
