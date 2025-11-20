module github.com/taufiq30s/chisa

go 1.24.0

toolchain go1.24.1

require (
	github.com/bwmarrin/discordgo v0.28.1
	github.com/disgoorg/disgolink/v3 v3.0.4
	github.com/disgoorg/snowflake/v2 v2.0.3
	github.com/go-co-op/gocron/v2 v2.16.1
	github.com/joho/godotenv v1.5.1
	github.com/redis/go-redis/v9 v9.7.3
)

replace github.com/disgoorg/disgolink/v3 => ./.vendor/disgolink

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/disgoorg/json v1.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
)
