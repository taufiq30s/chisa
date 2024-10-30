package bot

import (
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/redis/go-redis/v9"
	"github.com/taufiq30s/chisa/internal/music"
	"github.com/taufiq30s/chisa/internal/spotify"
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

		fmt.Println("Connecting to bot")
		utils.InfoLog.Println("Connecting to bot")
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
		fmt.Println("Create Cron Job")
		chisa.CreateJobs()
	}()

	fmt.Printf("Bot Ready with uptime: %s\n", time.Now().Format("Mon Jan 2 2006 15:04:05 GMT+0000"))
	utils.InfoLog.Printf("Bot Ready with uptime: %s\n", time.Now().Format("Mon Jan 2 2006 15:04:05 GMT+0000"))
}

func (bot *Bot) Disconnect() {
	err := bot.Session.Close()
	if err != nil {
		utils.ErrorLog.Fatalf("Failed to close connection: %s\n", err)
	}
	utils.InfoLog.Println("Bot connection closed")
}

func (bot *Bot) InitializeMusicClient(youtubeAPIKey string) {
	utils.InfoLog.Println("Connecting to music client")

	var wg sync.WaitGroup
	var spotifyClient *spotify.Client
	var lavalinkClient disgolink.Client
	var spotifyClientErr, lavalinkClientErr error

	wg.Add(2)

	go func() {
		defer wg.Done()
		spotifyClient, spotifyClientErr = spotify.New()
	}()

	go func() {
		defer wg.Done()
		lavalinkClient, lavalinkClientErr = music.NewLavalinkClient(bot.Session.State.User.ID)
	}()
	wg.Wait()

	if spotifyClientErr != nil {
		fmt.Printf("Failed to create spotify client: %s\n", spotifyClientErr)
		utils.ErrorLog.Fatalf("Failed to create spotify client: %s\n", spotifyClientErr)
	}

	if lavalinkClientErr != nil {
		fmt.Printf("Failed to create lavalink client: %s\n", lavalinkClientErr)
		utils.ErrorLog.Fatalf("Failed to create lavalink client: %s\n", lavalinkClientErr)
	}

	client := music.NewMusicClient(youtubeAPIKey, spotifyClient, lavalinkClient)
	bot.Music = music.NewMusicBot(client)

	fmt.Println("Music client connected")
	utils.InfoLog.Println("Music client connected")
}
