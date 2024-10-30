package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/internal/handlers"
	"github.com/taufiq30s/chisa/utils"
)

var chisa bot.Bot

func init() {
	fmt.Println("Chisa")
	err := godotenv.Load()
	if err != nil {
		utils.ErrorLog.Fatalf("Failed to load .env : %s\n", err)
	}
	chisa = bot.Bot{}
	go chisa.OpenRedis()
}

func main() {
	token, err := utils.GetEnv("BOT_TOKEN")
	if err != nil {
		utils.ErrorLog.Fatalln(err)
	}

	guildId, err := utils.GetEnv("AKASHIC_SERVER_ID")
	if err != nil {
		utils.ErrorLog.Fatalln(err)
	}

	youtubeAPIKey, err := utils.GetEnv("YOUTUBE_API_KEY")
	if err != nil {
		utils.ErrorLog.Fatalln(err)
	}

	// Initialize bot and start bot
	chisa.Start(token, guildId)

	// Initialize Spotify client
	go chisa.InitializeMusicClient(youtubeAPIKey)

	go handlers.Register(&chisa, guildId)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Unregister all commands
	go chisa.CloseRedis()
	defer chisa.Disconnect()
}
