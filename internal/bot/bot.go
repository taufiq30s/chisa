package bot

import (
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	"github.com/taufiq30s/chisa/internal/config"
	"github.com/taufiq30s/chisa/internal/currency"
	"github.com/taufiq30s/chisa/internal/music"
	"github.com/taufiq30s/chisa/utils"
)

type Bot struct {
	Config   *config.Config
	Session  *discordgo.Session
	Music    music.MusicService
	Currency currency.CurrencyService
	Redis    *redis.Client
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

		utils.InfoLog.Println("Connecting to discord...")
		err = session.Open()
		if err != nil {
			utils.ErrorLog.Fatalf("Failed to open connection: %s\n", err)
		}
		utils.InfoLog.Println("Bot connection open")

		chisa.Session = session
	}()
	wg.Wait()

	go func() {
		utils.InfoLog.Println("Creating Cron Job...")
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

	musicBot := music.New(
		bot.Session,
		bot.Session.State.User.ID,
		guildId,
	)

	if err := musicBot.ConnectToNodes(); err != nil {
		utils.ErrorLog.Printf("Music client failed to connect to nodes: %v — music commands will be unavailable\n", err)
	} else {
		musicBot.InitializeListenerFunctions()
		go musicBot.InitializeCleanSearchCache()
	}

	bot.Music = musicBot
}

func (bot *Bot) InitializeCurrencyClient(wg *sync.WaitGroup) {
	defer wg.Done()
	utils.InfoLog.Println("Connecting to currency client...")
	bot.Currency = currency.New(bot.Config.CurrencyProvider, bot.Config, bot.Redis)
}
