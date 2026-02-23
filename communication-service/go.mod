module github.com/tronget/human-touch/communication-service

go 1.25.4

require (
	github.com/go-chi/chi/v5 v5.2.5
	github.com/gorilla/websocket v1.5.3
	github.com/jmoiron/sqlx v1.4.0
	github.com/lib/pq v1.11.2
	github.com/tronget/human-touch/shared v0.0.0
)

replace github.com/tronget/human-touch/shared => ../shared
