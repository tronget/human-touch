package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var db *DB

func main() {
	dsn := os.Getenv("DATABASE_URL")
	dbx, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db = &DB{X: dbx}

	r := chi.NewRouter()
	// JWT middleware or gateway will pass user info
	r.Post("/messages", CreateMessage)
	r.Get("/messages/{convId}", GetMessages)
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
