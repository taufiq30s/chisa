package bot

import (
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron/v2"
	"github.com/redis/go-redis/v9"
	"github.com/taufiq30s/chisa/internal/config"
	"github.com/taufiq30s/chisa/internal/currency"
	"github.com/taufiq30s/chisa/internal/music"
	"github.com/taufiq30s/chisa/internal/quiz"
	"github.com/taufiq30s/chisa/utils"
)

type Bot struct {
	Config    *config.Config
	Session   *discordgo.Session
	Music     music.MusicService
	Currency  currency.CurrencyService
	Quiz      quiz.QuizService
	Redis     *redis.Client
	scheduler gocron.Scheduler
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

// StopJobs gracefully shuts down the cron scheduler.
func (bot *Bot) StopJobs() {
	if bot.scheduler != nil {
		if err := bot.scheduler.Shutdown(); err != nil {
			utils.ErrorLog.Printf("Error stopping cron scheduler: %v\n", err)
		} else {
			utils.InfoLog.Println("Cron scheduler stopped")
		}
	}
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

func (bot *Bot) InitializeQuizService(wg *sync.WaitGroup) {
	defer wg.Done()
	utils.InfoLog.Println("Initializing quiz service...")
	manager := quiz.New(bot.Config.QuizWSPort)
	quiz.SetDiscordSession(bot.Session)
	quiz.StartServer(manager, bot.Config.QuizWSPort)
	bot.Quiz = manager
	utils.InfoLog.Printf("Quiz service ready (WS port %s)\n", bot.Config.QuizWSPort)
}
