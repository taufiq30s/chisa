package music

type MusicTrack struct {
	Title     string
	Artist    string
	Length    int
	Thumbnail string
	Url       string
	AddedBy   string
}

type MusicBot struct {
	Client                   *MusicClient
	Queue                    []MusicTrack
	featureName              string
	supportedPlatformsPrefix map[string][]string
}

func NewMusicBot(client *MusicClient) *MusicBot {
	return &MusicBot{
		Client:      client,
		Queue:       make([]MusicTrack, 0),
		featureName: "Chisa Music Player",
		supportedPlatformsPrefix: map[string][]string{
			"youtube": {
				"https://youtube.com/",
				"https://www.youtube.com/",
				"https://www.youtube.com/shorts/",
				"https://youtu.be/",
			},
			"spotify": {
				"https://open.spotify.com",
			},
		},
	}
}
