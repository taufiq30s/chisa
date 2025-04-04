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
	var wg sync.WaitGroup
	wg.Add(3)

	token, err := utils.GetEnv("BOT_TOKEN")
	if err != nil {
		utils.ErrorLog.Fatalln(err)
	}

	guildId, err := utils.GetEnv("AKASHIC_SERVER_ID")
	if err != nil {
		utils.ErrorLog.Fatalln(err)
	}

	currencyAPI, err := utils.GetEnv("WISE_TOKEN")
	if err != nil {
		utils.ErrorLog.Fatalln(err)
	}

	// Initialize bot and start bot
	chisa.Start(token, guildId)

	// Initialize Feature and Handlers
	go chisa.InitializeMusicClient(&wg, guildId)
	go chisa.InitializeCurrencyClient(&wg, currencyAPI)
	go handlers.Register(&wg, &chisa, guildId)

	wg.Wait()
	fmt.Printf("Bot Ready with uptime: %s\n", time.Now().Format("Mon Jan 2 2006 15:04:05 GMT+0000"))
	utils.InfoLog.Printf("Bot Ready with uptime: %s\n", time.Now().Format("Mon Jan 2 2006 15:04:05 GMT+0000"))

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Unregister all commands
	go chisa.CloseRedis()
	go handlers.Unregister(&chisa, guildId)
	defer chisa.Disconnect()
}
