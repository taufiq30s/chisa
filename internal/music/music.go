package music

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/snowflake/v2"
	"github.com/taufiq30s/chisa/internal/responses"
	"github.com/taufiq30s/chisa/utils"
)

type MusicBot struct {
	Client        disgolink.Client
	queue         []trackState
	nodes         []disgolink.NodeConfig
	player        disgolink.Player
	searchResults map[string]*searchResultState
	session       *discordgo.Session
	lastPosition  int64
	featureName   string
}

func New(s *discordgo.Session, botId string, guildId string) *MusicBot {
	client := disgolink.New(
		snowflake.MustParse(botId),
	)
	player := client.Player(snowflake.MustParse(guildId))

	return &MusicBot{
		Client:        client,
		queue:         make([]trackState, InitialQueueCapacity),
		searchResults: make(map[string]*searchResultState),
		player:        player,
		session:       s,
		lastPosition:  InitialLastPosition,
		featureName:   "Chisa Music Player",
	}
}

func (m *MusicBot) InitializeListenerFunctions() {
	m.Client.AddListeners(
		disgolink.NewListenerFunc(m.onTrackEnd),
		disgolink.NewListenerFunc(m.onTrackStart),
		disgolink.NewListenerFunc(m.onTrackStuck),
		disgolink.NewListenerFunc(m.onWebSocketClosed),
		disgolink.NewListenerFunc(m.onTrackException),
		disgolink.NewListenerFunc(m.onPlayerUpdate),
	)
}

/*
Create a new lavalink client
*/
func (m *MusicBot) ConnectToNodes() error {
	var wg sync.WaitGroup
	err := m.loadNodes()

	if err != nil {
		utils.ErrorLog.Println("Failed to load lavalink configuration")
		return fmt.Errorf("failed to load lavalink configuration: %w", err)
	}

	for _, nodeConfiguration := range m.nodes {
		wg.Add(WaitGroupIncrement)
		go m.connectToNode(&nodeConfiguration, &wg)
	}
	wg.Wait()

	if m.Client.BestNode() == nil {
		fmt.Println("Music client failed to connected")
		utils.ErrorLog.Println("Music client failed to connected")
		return fmt.Errorf("music client failed to connected")
	}
	fmt.Println("Music client connected")
	utils.InfoLog.Println("Music client connected")
	return nil
}

func (m *MusicBot) connectToNode(config *disgolink.NodeConfig, wg *sync.WaitGroup) {
	fmt.Printf("Connecting to lavalink node %s\n", config.Name)
	ctx, cancel := context.WithTimeout(context.Background(), NodeConnectionTimeout)
	defer cancel()
	defer wg.Done()

	node, err := m.Client.AddNode(ctx, *config)
	if err != nil {
		utils.ErrorLog.Printf("Failed to connect to lavalink node %s\n", config.Name)
		fmt.Printf("Failed to connect to lavalink node %s\n", config.Name)
		return
	}

	// Get lavalink node version and make sure the node is healthy
	version, err := node.Version(ctx)
	if err != nil {
		utils.ErrorLog.Println("Failed to get lavalink node version")
		fmt.Printf("Failed to connect to lavalink node %s\n", config.Name)
		return
	}

	stats := node.Stats()
	if !checkNodeHealth(&stats) {
		fmt.Printf("Node %s is unhealthy\n", node.Config().Name)
		utils.ErrorLog.Printf("Node %s is unhealthy\n", node.Config().Name)
		m.Client.RemoveNode(node.Config().Name)
		return
	}

	utils.InfoLog.Printf("Connected to lavalink node: %s version: %s\n", node.Config().Name, version)
	fmt.Printf("Connected to lavalink node: %s version: %s\n", node.Config().Name, version)
}

/*
Load lavalink nodes from lavalink_clients.json
*/
func (m *MusicBot) loadNodes() error {
	nodesFile, err := os.Open("lavalink_nodes.json")
	if err != nil {
		return err
	}

	defer nodesFile.Close()
	bufferFile, _ := io.ReadAll(nodesFile)
	json.Unmarshal(bufferFile, &m.nodes)
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
				Value:  fmt.Sprintf("%02d:%02d", int(info.Length)/(MillisecondsPerSecond*SecondsPerMinute), int(info.Length)/MillisecondsPerSecond%SecondsPerMinute),
				Inline: true,
			},
			{
				Name:   "Provider",
				Value:  info.Provider,
				Inline: true,
			},
			{
				Name:   "Queue Length",
				Value:  fmt.Sprintf("%d", len(m.queue)-QueueOffsetForLength),
				Inline: true,
			},
		}),
	)
	if state.interaction != nil {
		err := responses.InteractionResponse(m.session, state.interaction.Interaction).WithEmbed(body).SendDefer()
		if err != nil {
			fmt.Println("Failed to send interaction response", err)
			utils.ErrorLog.Println("Failed to send interaction response", err)
		}
	} else {
		_, err := m.session.ChannelMessageSendEmbed(*state.channelId, body)
		if err != nil {
			fmt.Println("Failed to send message", err)
			utils.ErrorLog.Println("Failed to send message", err)
		}
	}
}
