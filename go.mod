module github.com/yourname/multiagent-social

go 1.24.0

toolchain go1.24.11

require (
	github.com/go-chi/chi/v5 v5.0.8
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/jackc/pgx/v5 v5.9.0
	github.com/pgvector/pgvector-go v0.4.0
	github.com/prometheus/client_golang v1.16.0
	github.com/redis/go-redis/v9 v9.0.0
	github.com/sashabaranov/go-openai v1.12.0
	nhooyr.io/websocket v1.10.2
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.10.1 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
)

replace github.com/jackc/pgx/v5 v5.9.0 => github.com/jackc/pgx/v5 v5.8.0

replace github.com/pgvector/pgvector-go v0.4.0 => github.com/pgvector/pgvector-go v0.3.0

replace nhooyr.io/websocket v1.10.2 => nhooyr.io/websocket v1.8.14
