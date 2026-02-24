module github.com/tronget/human-touch/auth-service

go 1.26.0

require (
	github.com/go-chi/chi/v5 v5.2.5
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/ilyakaznacheev/cleanenv v1.5.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/lib/pq v1.11.2
	github.com/tronget/human-touch/shared v0.0.0
	golang.org/x/crypto v0.48.0
)

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	olympos.io/encoding/edn v0.0.0-20201019073823-d3554ca0b0a3 // indirect
)

replace github.com/tronget/human-touch/shared => ../shared
