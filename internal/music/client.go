package music

import (
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/taufiq30s/chisa/internal/spotify"
)

type MusicClient struct {
	YoutubeAPIKey string
	SpotifyClient *spotify.Client
	Lavalink      disgolink.Client
}

func NewMusicClient(youtubeAPIKey string, spotifyClient *spotify.Client, lavalinkClient disgolink.Client) *MusicClient {
	return &MusicClient{
		YoutubeAPIKey: youtubeAPIKey,
		SpotifyClient: spotifyClient,
		Lavalink:      lavalinkClient,
	}
}
