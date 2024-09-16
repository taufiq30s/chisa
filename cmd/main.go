package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/internal/cronjob"
	"github.com/taufiq30s/chisa/internal/handlers"
	"github.com/taufiq30s/chisa/utils"
)

func init() {
	fmt.Println("Chisa")
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Failed to load .env : %s", err)
	}
	go bot.OpenRedis()
}

func main() {
	token, err := utils.GetEnv("BOT_TOKEN")
	if err != nil {
		log.Fatal(err)
	}

	guildId, err := utils.GetEnv("AKASHIC_SERVER_ID")
	if err != nil {
		log.Fatal(err)
	}

	chisa := bot.New()
	startBot(&chisa, token, guildId)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Unregister all commands
	go handlers.UnregisterCommands(&chisa, &guildId)
	go bot.CloseRedis()
	defer chisa.Disconnect()
}

func startBot(chisa *bot.Bot, token string, guildId string) {
	var wg sync.WaitGroup

	// Open bot connection and register handler
	wg.Add(1)
	isBotConnected := make(chan bool)
	go func() {
		defer wg.Done()
		chisa.Connect(token)
		isBotConnected <- true
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-isBotConnected
		fmt.Println("Initialized Commands and Events")
		handlers.RegisterCommands(chisa, &guildId)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		chisa.InitializeSpotifyClient()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Create Cron Job")
		cronjob.CreateJobs()
	}()

	// Wait untill all initialization process done
	// then send message
	wg.Wait()
	log.Printf("Bot Ready with uptime: %s", time.Now().Format("Mon Jan 2 2006 15:04:05 GMT+0000"))
	fmt.Println("Bot Ready")
}
