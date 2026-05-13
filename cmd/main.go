package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/internal/config"
	"github.com/taufiq30s/chisa/internal/handlers"
	"github.com/taufiq30s/chisa/internal/moderation"
	"github.com/taufiq30s/chisa/utils"
)

var chisa bot.Bot

func init() {
	fmt.Println("Chisa")
	err := godotenv.Load()
	if err != nil {
		utils.ErrorLog.Fatalf("Failed to load .env : %s\n", err)
	}

	cfg, err := config.Load()
	if err != nil {
		utils.ErrorLog.Fatalf("Configuration error: %s\n", err)
	}

	moderation.SetConfig(cfg)

	chisa = bot.Bot{Config: cfg}
	go chisa.OpenRedis()
}

func main() {
	var wg sync.WaitGroup
	wg.Add(NumInitGoroutines)

	// Initialize bot and start bot
	chisa.Start(chisa.Config.BotToken, chisa.Config.GuildID)

	// Initialize Feature and Handlers
	go chisa.InitializeMusicClient(&wg, chisa.Config.GuildID)
	go chisa.InitializeCurrencyClient(&wg)
	go handlers.Register(&wg, &chisa, chisa.Config.GuildID)

	wg.Wait()
	fmt.Printf("Bot Ready with uptime: %s\n", time.Now().Format(DateTimeFormat))
	utils.InfoLog.Printf("Bot Ready with uptime: %s\n", time.Now().Format(DateTimeFormat))

	sc := make(chan os.Signal, SignalChannelBuffer)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Unregister all commands
	go chisa.CloseRedis()
	go handlers.Unregister(&chisa, chisa.Config.GuildID)
	defer chisa.Disconnect()
}
