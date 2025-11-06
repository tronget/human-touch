package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://pguser:pgpass@localhost:5432/microdb?sslmode=disable"
	}
	dbx, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db := &DB{X: dbx}

	r := chi.NewRouter()
	r.Use(WithDB(db))

	r.Post("/register", RegisterHandler)
	r.Post("/login", LoginHandler)

	// protected example
	r.Group(func(r chi.Router) {
		r.Use(JWTAuth)
		r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
			uid := r.Context().Value(CtxUserID).(int64)
			// fetch user...
			_ = uid
			w.Write([]byte("ok"))
		})
	})

	port := os.Getenv("AUTH_SERVICE_PORT")
	if port == "" {
		port = "8001"
	}
	log.Println("auth service listening on", port)
	http.ListenAndServe(":"+port, r)
}
