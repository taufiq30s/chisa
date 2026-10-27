package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/taufiq30s/chisa/internal/model"
	"github.com/taufiq30s/chisa/utils"
)

const YOUTUBE_URL_API_SEARCH = "https://www.googleapis.com/youtube/v3/search"
const YOUTUBE_URL_API_VIDEOS = "https://www.googleapis.com/youtube/v3/videos"

// Youtube API V3 Search Response Structs
type SearchListResponse struct {
	Items []SearchResultItem `json:"items"`
}

type SearchResultItem struct {
	ID      SearchResultID `json:"id"`
	Snippet SearchSnippet  `json:"snippet"`
}

type SearchResultID struct {
	Kind    string `json:"kind"`
	VideoID string `json:"videoId"`
}

type SearchSnippet struct {
	Title        string `json:"title"`
	ChannelTitle string `json:"channelTitle"`
}

// Youtube API V3 Video Response Structs
type VideoListResponse struct {
	Items []VideoItem `json:"items"`
}

type VideoItem struct {
	ID             string         `json:"id"`
	Snippet        Snippet        `json:"snippet"`
	ContentDetails ContentDetails `json:"contentDetails"`
}

type Snippet struct {
	Title        string     `json:"title"`
	ChannelTitle string     `json:"channelTitle"`
	Thumbnails   Thumbnails `json:"thumbnails"`
}

type Thumbnails struct {
	Default  Thumbnail `json:"default"`
	Medium   Thumbnail `json:"medium"`
	High     Thumbnail `json:"high"`
	Standard Thumbnail `json:"standard"`
	Maxres   Thumbnail `json:"maxres"`
}

type Thumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type ContentDetails struct {
	Duration string `json:"duration"` // ISO 8601, e.g. "PT5M25S"
}

// YoutubeProvider struct

type YoutubeProvider struct {
	apiKey     string
	httpClient *http.Client
}

func NewYoutubeProvider() (*YoutubeProvider, error) {
	apiKey, err := utils.GetEnv("YOUTUBE_API_KEY")
	if err != nil {
		return nil, err
	}
	return &YoutubeProvider{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}, nil
}

// SearchYouTube performs a search query on YouTube and returns the search results.
//
// Parameters:
//   - query: The search query string to be sent to the YouTube API.
//   - pageToken: An optional token for pagination to retrieve the next set of results.
//
// Returns:
//   - A slice of SearchResultItem structs containing the search results.
//   - An error if the search operation fails.
func (p *YoutubeProvider) searchYouTube(query string, pageToken string) ([]SearchResultItem, error) {
	req, err := http.NewRequest("GET", YOUTUBE_URL_API_SEARCH, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("part", "snippet")
	q.Add("q", query)
	q.Add("maxResults", "25")
	q.Add("type", "video")
	q.Add("key", p.apiKey)
	if pageToken != "" {
		q.Add("pageToken", pageToken)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var searchResponse SearchListResponse
	err = json.NewDecoder(resp.Body).Decode(&searchResponse)
	if err != nil {
		return nil, err
	}

	return searchResponse.Items, nil
}

// videoIDBatchSize is the API maximum for a single videos.list call. One call
// with 50 IDs costs the same 1 unit as one call with a single ID, so batching is
// what keeps the fallback path cheap.
const videoIDBatchSize = 50

// getVideoDetails fetches snippet + duration for up to 50 IDs per request.
//
// Returns a map keyed by video ID: videos.list silently omits IDs that are
// deleted, private, or region-blocked, and callers must not assume 1:1 ordering.
func (p *YoutubeProvider) getVideoDetails(videoIDs []string) (map[string]VideoItem, error) {
	details := make(map[string]VideoItem, len(videoIDs))

	for start := 0; start < len(videoIDs); start += videoIDBatchSize {
		end := min(start+videoIDBatchSize, len(videoIDs))
		if err := p.fetchVideoBatch(videoIDs[start:end], details); err != nil {
			return nil, err
		}
	}
	return details, nil
}

func (p *YoutubeProvider) fetchVideoBatch(videoIDs []string, into map[string]VideoItem) error {
	req, err := http.NewRequest("GET", YOUTUBE_URL_API_VIDEOS, nil)
	if err != nil {
		return err
	}
	q := req.URL.Query()
	q.Add("part", "snippet,contentDetails")
	q.Add("id", strings.Join(videoIDs, ","))
	q.Add("key", p.apiKey)
	req.URL.RawQuery = q.Encode()

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var videoResponse VideoListResponse
	if err := json.NewDecoder(resp.Body).Decode(&videoResponse); err != nil {
		return err
	}
	for _, item := range videoResponse.Items {
		into[item.ID] = item
	}
	return nil
}

// Accepts checks if the given query is a YouTube URL or a search query.
//
// Parameters:
//   - query: The query string to be checked.
//
// Returns:
//   - A boolean indicating whether the query is accepted by the YouTube provider.
func (p *YoutubeProvider) Accepts(query string) bool {
	if !IsURL(query) {
		return true
	}
	return strings.HasPrefix(query, "https://www.youtube.com") ||
		strings.HasPrefix(query, "https://youtu.be") ||
		strings.HasPrefix(query, "https://youtube.com")
}

// Resolve resolves the given query into a list of tracks.
//
// Backend order: yt-dlp first, YouTube Data API only as a fallback. yt-dlp costs
// zero quota units and returns the duration that search.list omits; the Data API
// path exists so a host without yt-dlp still works.
//
// Parameters:
//   - ctx: The context for the operation.
//   - query: The query string to be resolved (a YouTube URL or a text search).
//
// Returns:
//   - A slice of Track structs representing the resolved tracks.
//   - A boolean indicating whether the result is a search result (the caller
//     shows a paginated picker) rather than a single track to enqueue.
//   - An error if the resolution operation fails.
func (p *YoutubeProvider) Resolve(ctx context.Context, query string) ([]model.Track, bool, error) {
	if !IsURL(query) {
		tracks, err := searchViaYtDlp(ctx, query, ytdlpSearchLimit)
		if err == nil {
			return tracks, true, nil
		}
		logYtdlpFallback(err)
		return p.searchViaAPI(query)
	}

	// A playlist expands to many tracks and is still not a search: the caller
	// must queue every result instead of showing the picker, so isSearch stays
	// false. That is how the caller tells a playlist from a single video.
	if isPlaylistURL(query) {
		tracks, err := playlistViaYtDlp(ctx, query, ytdlpPlaylistLimit)
		if err != nil {
			return nil, false, err
		}
		// A readable playlist can still be empty: every entry was a live stream
		// or unavailable. Saying so beats the caller's generic "check your URL".
		if len(tracks) == 0 {
			return nil, false, errors.New("playlist has no playable videos")
		}
		return tracks, false, nil
	}

	track, err := resolveURLViaYtDlp(ctx, query)
	if err == nil {
		return []model.Track{track}, false, nil
	}
	logYtdlpFallback(err)
	return p.resolveURLViaAPI(query)
}

// IsURL reports whether query parses as an absolute URI. Anything else is a text
// search, including strings that merely contain a URL. Exported because the
// Resolver needs the same rule to decide whether a result is cacheable.
func IsURL(query string) bool {
	_, err := url.ParseRequestURI(query)
	return err == nil
}

// isPlaylistURL reports whether query points at a playlist rather than a video.
//
// Only /playlist and /videoseries count. A "list=" parameter alone is not
// enough: "watch?v=X&list=Y" is one video inside a playlist, and the user who
// pasted it asked for that video, so v= wins and it resolves as a single track.
// youtu.be links can never carry a playlist, so they never match.
func isPlaylistURL(query string) bool {
	if !strings.Contains(query, "youtube.com") {
		return false
	}
	parsed, err := url.Parse(query)
	if err != nil {
		return false
	}
	if parsed.Query().Get("v") != "" {
		return false
	}
	return (parsed.Path == "/playlist" || parsed.Path == "/videoseries") &&
		parsed.Query().Get("list") != ""
}

// searchViaAPI is the fallback search path: search.list then one batched
// videos.list. Cost: 1 search-unit + 1 unit, and search.list is capped at 100
// calls/day, which is why it is not the primary path.
func (p *YoutubeProvider) searchViaAPI(query string) ([]model.Track, bool, error) {
	searchResults, err := p.searchYouTube(query, "")
	if err != nil {
		return nil, true, err
	}

	ids := make([]string, 0, len(searchResults))
	for _, item := range searchResults {
		if item.ID.VideoID != "" {
			ids = append(ids, item.ID.VideoID)
		}
	}
	if len(ids) == 0 {
		return []model.Track{}, true, nil
	}

	details, err := p.getVideoDetails(ids)
	if err != nil {
		return nil, true, err
	}

	tracks := make([]model.Track, 0, len(ids))
	for _, id := range ids {
		video, ok := details[id]
		if !ok {
			continue // deleted, private, or region-blocked
		}
		track, err := p.trackFromVideoItem(video)
		if err != nil {
			continue // unparsable duration: drop the entry, keep the search usable
		}
		tracks = append(tracks, track)
	}
	return tracks, true, nil
}

// resolveURLViaAPI is the fallback for a single video URL.
func (p *YoutubeProvider) resolveURLViaAPI(query string) ([]model.Track, bool, error) {
	videoID, err := videoIDFromURL(query)
	if err != nil {
		return nil, false, err
	}

	details, err := p.getVideoDetails([]string{videoID})
	if err != nil {
		return nil, false, err
	}
	video, ok := details[videoID]
	if !ok {
		return nil, false, fmt.Errorf("video not found: %s", videoID)
	}

	track, err := p.trackFromVideoItem(video)
	if err != nil {
		return nil, false, err
	}
	track.Url = query
	return []model.Track{track}, false, nil
}

// trackFromVideoItem maps an API item to a Track. Search results used to hardcode
// Length: 0, Url: "" and Thumbnail: "" - the first showed 00:00 and the empty Url
// made the track unstreamable (the player pipes Track.Url into yt-dlp).
func (p *YoutubeProvider) trackFromVideoItem(video VideoItem) (model.Track, error) {
	length, err := utils.FormatISO8601ToSeconds(video.ContentDetails.Duration)
	if err != nil {
		return model.Track{}, err
	}
	thumbnail := video.Snippet.Thumbnails.High.URL
	if thumbnail == "" {
		thumbnail = video.Snippet.Thumbnails.Default.URL
	}
	if thumbnail == "" {
		thumbnail = fmt.Sprintf("https://i.ytimg.com/vi/%s/hqdefault.jpg", video.ID)
	}
	return model.Track{
		ID:        video.ID,
		Title:     video.Snippet.Title,
		Artist:    video.Snippet.ChannelTitle,
		Length:    length,
		Thumbnail: thumbnail,
		Url:       "https://www.youtube.com/watch?v=" + video.ID,
		Provider:  "YouTube",
	}, nil
}

// videoIDFromURL extracts the 11-char video ID from watch, youtu.be, and /shorts
// URLs.
func videoIDFromURL(query string) (string, error) {
	parsedURL, err := url.Parse(query)
	if err != nil {
		return "", err
	}
	if videoID := parsedURL.Query().Get("v"); videoID != "" {
		return videoID, nil
	}
	if strings.Contains(parsedURL.Host, "youtu.be") {
		return strings.TrimPrefix(parsedURL.Path, "/"), nil
	}
	if strings.Contains(parsedURL.Path, "/shorts/") {
		return strings.TrimPrefix(parsedURL.Path, "/shorts/"), nil
	}
	return "", fmt.Errorf("invalid YouTube URL: %s", query)
}
