module github.com/tronget/human-touch/api-gateway

go 1.26.0

require (
	github.com/go-chi/chi/v5 v5.2.5
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/jmoiron/sqlx v1.4.0
	github.com/tronget/human-touch/shared v0.0.0
)

require github.com/lib/pq v1.11.2 // indirect

replace github.com/tronget/human-touch/shared => ../shared
