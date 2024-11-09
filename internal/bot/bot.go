package bot

import (
	"fmt"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	"github.com/taufiq30s/chisa/internal/music"
	"github.com/taufiq30s/chisa/utils"
)

type Bot struct {
	Session *discordgo.Session
	Music   *music.MusicBot
	Redis   *redis.Client
}

func (chisa *Bot) Start(token string, guildId string) {
	var wg sync.WaitGroup
	wg.Add(1)

	// Open bot connection and register handler
	go func() {
		defer wg.Done()

		session, err := discordgo.New("Bot " + token)
		if err != nil {
			utils.ErrorLog.Fatalf("Failed to created discord session: %s\n", err)
		}

		fmt.Println("Connecting to discord...")
		utils.InfoLog.Println("Connecting to discord...")
		err = session.Open()
		if err != nil {
			fmt.Printf("Failed to open connection: %s\n", err)
			utils.ErrorLog.Fatalf("Failed to open connection: %s\n", err)
		}
		fmt.Println("Bot connection open")
		utils.InfoLog.Println("Bot connection open")

		chisa.Session = session
	}()
	wg.Wait()

	go func() {
		fmt.Println("Creating Cron Job...")
		chisa.CreateJobs()
	}()
}

func (bot *Bot) Disconnect() {
	err := bot.Session.Close()
	if err != nil {
		utils.ErrorLog.Fatalf("Failed to close connection: %s\n", err)
	}
	utils.InfoLog.Println("Bot connection closed")
}

func (bot *Bot) InitializeMusicClient(wg *sync.WaitGroup, guildId string) {
	defer wg.Done()
	utils.InfoLog.Println("Connecting to music client...")
	fmt.Println("Connecting to music client...")

	musicBot := music.New(
		bot.Session,
		bot.Session.State.User.ID,
		guildId,
	)
	musicBot.ConnectToNodes()
	musicBot.InitializeListenerFunctions()
	go musicBot.InitializeCleanSearchCache()
	bot.Music = musicBot
}
