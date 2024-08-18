package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/internal/commands"
	"github.com/taufiq30s/chisa/utils"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Failed to load .env : %s", err)
	}
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
	chisa.Connect(token)
	chisa.InitializeSpotifyClient()

	fmt.Println("Chisa")

	fmt.Println("Initialized Commands and Events")
	commands.Register(&chisa, &guildId)

	fmt.Println("Bot Started")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	defer chisa.Disconnect()
}
