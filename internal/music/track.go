package music

import "github.com/disgoorg/disgolink/v3/lavalink"

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
	info := trackInfo{
		Title:    track.Info.Title,
		Artist:   track.Info.Author,
		Length:   int(track.Info.Length.Milliseconds()),
		Provider: track.Info.SourceName,
		AddedBy:  username,
	}
	if track.Info.ArtworkURL != nil {
		info.Thumbnail = *track.Info.ArtworkURL
	}
	if track.Info.URI != nil {
		info.Url = *track.Info.URI
	}
	return info
}
