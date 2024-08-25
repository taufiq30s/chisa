package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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
	bot.OpenRedis()
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

	// Open Bot and Spotify connection
	chisa := bot.New()
	chisa.Connect(token)
	chisa.InitializeSpotifyClient()

	fmt.Println("Initialized Commands and Events")
	handlers.Register(&chisa, &guildId)

	fmt.Println("Create Cron Job")
	cronjob.CreateJobs()

	// fmt.Println(chisa.Session.State.User.)
	fmt.Println("Bot Started")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	bot.CloseRedis()
	defer chisa.Disconnect()
}
