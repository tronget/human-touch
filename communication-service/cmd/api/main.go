package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/tronget/human-touch/communication-service/internal/handlers"
	"github.com/tronget/human-touch/communication-service/internal/middlewares"
	"github.com/tronget/human-touch/communication-service/internal/ws"
	"github.com/tronget/human-touch/shared/storage"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("Missing `DATABASE_URL` in environment variables")
	}
	dbx, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer dbx.Close()
	db := storage.NewDB(dbx, slog.Default())

	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(middlewares.WithDB(db), middlewares.WithUser(db))
		r.Get("/responses/chats/sent", handlers.GetChatsWhereUserIsSender)
		r.Get("/responses/chats/owned", handlers.GetChatsWhereUserIsOwner)
		r.Post("/responses/{responseId}/messages", handlers.CreateMessage)
		r.Get("/responses/{responseId}/messages", handlers.GetMessages)
		r.Get("/ws", ws.WsHandler)
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	port := os.Getenv("COMM_SERVICE_PORT")
	if port == "" {
		port = "8002"
	}
	log.Println("comm service on", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Printf("error: %v", err)
	}
}
