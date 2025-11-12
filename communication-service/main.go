package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/tronget/communication-service/storage"
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
	dbLogger := log.New(os.Stdout, "db ", log.LstdFlags|log.Lmicroseconds)
	db := storage.NewDB(dbx, dbLogger)

	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(WithDB(db), WithUID)
		r.Post("/messages", CreateMessage)
		r.Get("/messages/{convId}", GetMessages)
	})

	r.Get("/ws", wsHandler)

	port := os.Getenv("COMM_SERVICE_PORT")
	if port == "" {
		port = "8002"
	}
	log.Println("comm service on", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Printf("error: %v", err)
	}
}
