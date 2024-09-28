package music

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/kkdai/youtube/v2"
)

var client youtube.Client

func init() {
	client = youtube.Client{}
}

func getInfo(videoId string) (*youtube.Video, error) {
	return client.GetVideo(videoId)
}

func getAudio(video *youtube.Video) (*bytes.Reader, error) {
	formats := video.Formats.WithAudioChannels().Quality("tiny")
	stream, _, err := client.GetStream(video, &formats[0])
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	file, err := io.ReadAll(stream)
	if err != nil {
		return nil, err
	}
	fileReader := bytes.NewReader(file)

	return fileReader, nil
}

func getVideoId(url string) (string, error) {
	// Regular expression pattern to match the video ID
	var pattern string
	if strings.Contains(url, "youtube.com/watch?v=") {
		pattern = `v=([a-zA-Z0-9_-]+)`
	} else if strings.Contains(url, "youtube.com/shorts/") {
		pattern = `shorts/([a-zA-Z0-9_-]+)`
	} else if strings.Contains(url, "youtu.be/") {
		pattern = `youtu.be/([a-zA-Z0-9_-]+)`
	}
	// Compile the regular expression pattern
	re := regexp.MustCompile(pattern)

	// Find the video ID in the URL
	matches := re.FindStringSubmatch(url)

	// Check if the video ID was found
	if len(matches) < 2 {
		return "", fmt.Errorf("video ID not found in URL")
	}
	return matches[1], nil
}
