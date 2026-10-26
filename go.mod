module github.com/taufiq30s/chisa

go 1.27

require (
	github.com/bwmarrin/discordgo v0.29.0
	github.com/disgoorg/snowflake/v2 v2.0.3
	github.com/go-co-op/gocron/v2 v2.22.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
	github.com/joho/godotenv v1.5.1
	github.com/jonas747/ogg v0.0.0-20161220051205-b4f6f4cf3757
	github.com/redis/go-redis/v9 v9.23.0
	github.com/sony/gobreaker v1.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudflare/circl v1.6.5 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	go.uber.org/atomic v1.12.0 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
)

replace github.com/bwmarrin/discordgo => github.com/yeongaori/discordgo v0.0.0-20260321152711-3d3293e4c765
