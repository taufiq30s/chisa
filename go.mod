module github.com/taufiq30s/chisa

go 1.25.0

require (
	github.com/bwmarrin/discordgo v0.29.0
	github.com/disgoorg/disgolink/v3 v3.1.0
	github.com/disgoorg/snowflake/v2 v2.0.3
	github.com/fogleman/gg v1.3.0
	github.com/go-co-op/gocron/v2 v2.19.1
	github.com/joho/godotenv v1.5.1
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/redis/go-redis/v9 v9.18.0
	golang.org/x/image v0.37.0
)

replace github.com/disgoorg/disgolink/v3 => ./.vendor/disgolink

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/disgoorg/json v1.2.0 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/sony/gobreaker v1.0.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/crypto v0.49.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
)
