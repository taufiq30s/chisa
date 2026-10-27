package main

import (
	"context"
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
	"github.com/taufiq30s/chisa/internal/music"
	"github.com/taufiq30s/chisa/internal/responses"
	"github.com/taufiq30s/chisa/utils"
)

func main() {
	fmt.Println("Chisa")

	if err := godotenv.Load(); err != nil {
		utils.ErrorLog.Fatalf("Failed to load .env : %s\n", err)
	}

	cfg, err := config.Load()
	if err != nil {
		utils.ErrorLog.Fatalf("Configuration error: %s\n", err)
	}
	fmt.Println("Version:", cfg.Version)

	moderation.SetConfig(cfg)
	responses.SetVersion(cfg.Version)

	chisa := &bot.Bot{Config: cfg}
	chisa.OpenRedis()

	var wg sync.WaitGroup
	wg.Add(NumInitGoroutines)

	// Connect to Discord and start cron jobs.
	chisa.Start(cfg.BotToken, cfg.GuildID)
	moderation.SetRedis(chisa.Redis)
	go moderation.ReplayPendingVerifications(chisa.Session)

	// Initialize features and register Discord handlers.
	registry := handlers.NewRegistry(chisa)
	go chisa.InitializeMusicClient(&wg, cfg.GuildID)
	go chisa.InitializeCurrencyClient(&wg)
	go chisa.InitializeQuizService(&wg)
	go registry.Register(&wg, cfg.GuildID)

	wg.Wait()
	fmt.Printf("Bot Ready with uptime: %s\n", time.Now().Format(DateTimeFormat))
	utils.InfoLog.Printf("Bot Ready with uptime: %s\n", time.Now().Format(DateTimeFormat))

	sc := make(chan os.Signal, SignalChannelBuffer)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Graceful shutdown with timeout.
	_, shutdownCancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer shutdownCancel()

	utils.InfoLog.Println("Shutting down...")

	// Shutdown music bot gracefully
	if chisa.Music != nil {
		if musicBot, ok := chisa.Music.(*music.MusicBot); ok {
			musicBot.Shutdown()
		}
	}

	chisa.StopJobs()
	go chisa.CloseRedis()
	go registry.Unregister(cfg.GuildID)
	defer chisa.Disconnect()
}
