package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/internal/commands"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Failed to load .env : %s", err)
	}
}

func getEnv(key string) (string, error) {
	data, found := os.LookupEnv(key)
	if !found {
		return "", errors.New("key not found")
	}
	return data, nil
}

func main() {
	token, err := getEnv("BOT_TOKEN")
	if err != nil {
		log.Fatal(err)
	}

	guildId, err := getEnv("AKASHIC_SERVER_ID")
	if err != nil {
		log.Fatal(err)
	}

	chisa := bot.New()
	chisa.Connect(token)

	fmt.Println("Chisa")

	fmt.Println("Initialized Commands and Events")
	commands.Register(chisa.Session, &guildId)

	fmt.Println("Bot Started")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

}
