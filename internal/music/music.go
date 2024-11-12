package music

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
	"github.com/taufiq30s/chisa/internal/responses"
	"github.com/taufiq30s/chisa/utils"
)

type trackInfo struct {
	Title     string
	Artist    string
	Length    int
	Thumbnail string
	Url       string
	Provider  string
	AddedBy   *string
}

func newMusicTrack(track *lavalink.Track, username *string) trackInfo {
	return trackInfo{
		Title:     track.Info.Title,
		Artist:    track.Info.Author,
		Length:    int(track.Info.Length.Milliseconds()),
		Thumbnail: *track.Info.ArtworkURL,
		Url:       *track.Info.URI,
		Provider:  track.Info.SourceName,
		AddedBy:   username,
	}
}

type MusicBot struct {
	Client        disgolink.Client
	queue         []trackState
	nodes         []disgolink.NodeConfig
	player        disgolink.Player
	searchResults map[string]*searchResultState
	session       *discordgo.Session
	featureName   string
}

func New(s *discordgo.Session, botId string, guildId string) *MusicBot {
	client := disgolink.New(
		snowflake.MustParse(botId),
	)
	player := client.Player(snowflake.MustParse(guildId))

	return &MusicBot{
		Client:        client,
		queue:         make([]trackState, 0),
		searchResults: make(map[string]*searchResultState),
		player:        player,
		session:       s,
		featureName:   "Chisa Music Player",
	}
}

func (m *MusicBot) InitializeListenerFunctions() {
	m.Client.AddListeners(
		disgolink.NewListenerFunc(m.onTrackEnd),
		disgolink.NewListenerFunc(m.onTrackStart),
		disgolink.NewListenerFunc(m.onTrackStuck),
		disgolink.NewListenerFunc(m.onWebSocketClosed),
	)
}

/*
Create a new lavalink client
*/
func (m *MusicBot) ConnectToNodes() error {
	utils.InfoLog.Println("Connecting to lavalink nodes")
	fmt.Println("Connecting to lavalink nodes")
	var wg sync.WaitGroup
	err := m.loadNodes()

	if err != nil {
		utils.ErrorLog.Println("Failed to load lavalink configuration")
		return fmt.Errorf("failed to load lavalink configuration: %w", err)
	}

	wg.Add(len(m.nodes))
	for _, nodeConfiguration := range m.nodes {
		go m.connectToNode(&nodeConfiguration, &wg)
	}
	wg.Wait()

	if m.Client.BestNode() == nil {
		fmt.Println("Music client failed to connected")
		utils.ErrorLog.Println("Music client failed to connected")
	}
	fmt.Println("Music client connected")
	utils.InfoLog.Println("Music client connected")
	return nil
}

func (m *MusicBot) connectToNode(config *disgolink.NodeConfig, wg *sync.WaitGroup) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer wg.Done()

	node, err := m.Client.AddNode(ctx, *config)
	if err != nil {
		utils.ErrorLog.Printf("Failed to connect to lavalink node %s\n", config.Name)
		fmt.Printf("Failed to connect to lavalink node %s\n", config.Name)
		return
	}

	version, err := node.Version(ctx)
	if err != nil {
		utils.ErrorLog.Println("Failed to get lavalink node version")
		fmt.Printf("Failed to connect to lavalink node %s\n", config.Name)
		return
	}
	utils.InfoLog.Printf("Connected to lavalink node: %s version: %s\n", node.Config().Name, version)
	fmt.Printf("Connected to lavalink node: %s version: %s\n", node.Config().Name, version)
}

/*
Load lavalink nodes from lavalink_clients.json
*/
func (MusicBot *MusicBot) loadNodes() error {
	nodesFile, err := os.Open("lavalink_nodes.json")
	if err != nil {
		return err
	}

	defer nodesFile.Close()
	bufferFile, _ := io.ReadAll(nodesFile)
	json.Unmarshal(bufferFile, &MusicBot.nodes)
	return nil
}

func (m *MusicBot) generateMusicInformation(state *trackState, status string) {
	info := newMusicTrack(state.lavaTrack, state.addedBy)
	if state.playlistName != nil {
		info.Title = *state.playlistName
	}

	body := responses.CreateMessageEmbed(
		m.session, info.Title,
		"",
		fmt.Sprintf("%s (%s)", m.featureName, status),
		responses.SetThumbnailUrl(info.Thumbnail),
		responses.SetUrl(info.Url),
		responses.SetColor("5e11d9"),
		responses.SetFields([]*discordgo.MessageEmbedField{
			{
				Name:   "Added by",
				Value:  *info.AddedBy,
				Inline: true,
			},
			{
				Name:   "Duration",
				Value:  fmt.Sprintf("%02d:%02d", int(info.Length)/(1000*60), int(info.Length)/1000%60),
				Inline: true,
			},
			{
				Name:   "Provider",
				Value:  info.Provider,
				Inline: true,
			},
			{
				Name:   "Queue Length",
				Value:  fmt.Sprintf("%d", len(m.queue)-1),
				Inline: true,
			},
		}),
	)
	if state.interaction != nil {
		responses.InteractionResponse(m.session, state.interaction.Interaction,
			discordgo.InteractionResponseChannelMessageWithSource,
			false,
			body,
		).Execute()
	} else {
		m.session.ChannelMessageSendEmbed(*state.channelId, body)
	}
}
