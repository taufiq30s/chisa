package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/taufiq30s/chisa/internal/model"
	"github.com/taufiq30s/chisa/utils"
)

// ytdlpSearchLimit is how many entries we ask yt-dlp for. It is deliberately
// larger than music.MaxSearchResults (20): the filter below drops live streams
// and premieres, so asking for exactly 20 would sometimes page to 19 results.
const ytdlpSearchLimit = 25

// ytdlpPlaylistLimit caps one playlist import. 100 entries is a single flat
// extract (one subprocess, no per-video request), so this only bounds memory and
// the summary count, not the number of yt-dlp runs.
const ytdlpPlaylistLimit = 100

// ErrPlaylistUnavailable is returned for a playlist that cannot be read. YouTube
// reports private and nonexistent playlists with the same message, so the two
// cases cannot be told apart and share one user-facing error.
var ErrPlaylistUnavailable = errors.New(
	"playlist cannot be played because it is private or does not exist, check the URL and try again")

// FindYtdlpBinary returns the yt-dlp executable to run: the copy shipped in the
// repo is preferred so a deployment is not at the mercy of whatever the host has
// installed. Falls back to "yt-dlp" on PATH.
//
// Lives here (not on Streamer) because both playback and search need it, and
// search must not depend on a live voice connection.
func FindYtdlpBinary() string {
	// YTDLP_PATH wins so a host can point at its own build (a newer release, or a
	// wrapper script) without replacing the bundled binary. Checked first because
	// the bundled copy is a fallback, not an override.
	if custom, err := utils.GetEnv("YTDLP_PATH"); err == nil && custom != "" {
		return custom
	}
	localPath := filepath.Join(".", "bin", "yt-dlp")
	if _, err := os.Stat(localPath); err == nil {
		return localPath
	}
	return "yt-dlp"
}

// CommonArgs are the flags every invocation needs. They live in one place so
// search and playback cannot drift apart: search builds the watch URL that
// playback later fetches, and a PO token obtained by one is reusable by the
// other only if both ask for it.
//
// The bgutil PO token provider is what stops YouTube serving 403 on playback.
// yt-dlp defaults the plugin to http://127.0.0.1:4416, which is never right in a
// container, so the URL is explicit. YOUTUBE_POT_URL is optional: when it is
// unset the flag is omitted and yt-dlp falls back to its own PO token handling,
// which is enough for search but usually fails for playback.
func CommonArgs() []string {
	args := []string{"--no-warnings"}
	if potURL, err := utils.GetEnv("YOUTUBE_POT_URL"); err == nil && potURL != "" {
		args = append(args, "--extractor-args", "youtubepot-bgutilhttp:base_url="+potURL)
	}
	return args
}

// ytdlpEntry is one item of yt-dlp's JSON. Only the fields we consume are
// declared; yt-dlp emits dozens more and json ignores them.
type ytdlpEntry struct {
	// Type is "url" in --flat-playlist mode (search) and "video" in full-extract
	// mode (single URL), so both are treated as a single playable item.
	Type string `json:"_type"`
	// IEKey and ExtractorKey carry the same information under different names
	// depending on mode: flat listings set ie_key, full extraction sets
	// extractor_key. Neither is present in both.
	IEKey        string           `json:"ie_key"`
	ExtractorKey string           `json:"extractor_key"`
	ID           string           `json:"id"`
	Title        string           `json:"title"`
	Duration     *float64         `json:"duration"` // POINTER: nil means "not a finite video" (live/premiere)
	Thumbnails   []ytdlpThumbnail `json:"thumbnails"`
	URL          string           `json:"url"`         // watch URL in flat mode
	WebpageURL   string           `json:"webpage_url"` // absolute URL in full-extract mode
	Channel      string           `json:"channel"`
}

type ytdlpThumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type ytdlpPlaylist struct {
	Entries []ytdlpEntry `json:"entries"`
}

// keepVideoEntry reports whether an entry is a playable, finite video.
//
// Filtering here (not repairing later) is what enforces "video only": each nil
// duration would otherwise reach the embed as 00:00 and reach the player as a
// stream that never ends. Each check catches a distinct case:
//   - Type is neither "url" nor "video": a playlist/channel hit, not one item.
//   - No YouTube extractor key: the entry came from a different extractor.
//   - ID length != 11: YouTube video IDs are exactly 11 chars; playlist IDs are not.
//   - Duration nil or 0: live, upcoming premiere, or an unknown length.
//
// The type and extractor key have two spellings because yt-dlp reports the same
// item differently in the two modes we run:
//
//	--flat-playlist (search): _type="url",   ie_key="Youtube",        no extractor_key
//	full extract   (one URL): _type="video", extractor_key="Youtube", no ie_key
//
// Checking only the flat shape silently rejects every direct URL, which is why
// this must stay tolerant of both.
func keepVideoEntry(e ytdlpEntry) bool {
	isVideo := e.Type == "url" || e.Type == "video"
	isYouTube := e.IEKey == "Youtube" || e.ExtractorKey == "Youtube"
	return isVideo &&
		isYouTube &&
		len(e.ID) == 11 &&
		e.Duration != nil && *e.Duration > 0
}

// pickThumbnail returns the widest thumbnail, or the well-known hqdefault URL
// when yt-dlp gave none. --flat-playlist only reports 1-2 sizes, so "widest
// available" is the best we can do without a second request per video.
func pickThumbnail(e ytdlpEntry) string {
	best := ""
	bestWidth := -1
	for _, t := range e.Thumbnails {
		if t.URL != "" && t.Width > bestWidth {
			best, bestWidth = t.URL, t.Width
		}
	}
	if best != "" {
		return best
	}
	return fmt.Sprintf("https://i.ytimg.com/vi/%s/hqdefault.jpg", e.ID)
}

// trackFromEntry converts a passing entry. The watch URL is rebuilt from the ID
// when yt-dlp did not give one, because the streamer pipes track.Url into yt-dlp
// and an empty URL means an unplayable queue item.
func trackFromEntry(e ytdlpEntry) model.Track {
	watchURL := e.URL
	if watchURL == "" {
		watchURL = e.WebpageURL
	}
	if watchURL == "" {
		watchURL = "https://www.youtube.com/watch?v=" + e.ID
	}
	return model.Track{
		ID:        e.ID,
		Title:     e.Title,
		Artist:    e.Channel,
		Length:    int(*e.Duration + 0.5), // yt-dlp reports seconds as a float
		Thumbnail: pickThumbnail(e),
		Url:       watchURL,
		Provider:  "YouTube",
	}
}

// runYtdlp executes yt-dlp and returns stdout. stderr is only surfaced on error:
// yt-dlp writes Python deprecation warnings there on success, and mixing them
// into stdout would break the JSON decode.
func runYtdlp(ctx context.Context, args ...string) ([]byte, error) {
	full := append(CommonArgs(), args...)
	cmd := exec.CommandContext(ctx, FindYtdlpBinary(), full...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("yt-dlp %v: %w: %s", full, err, stderr.String())
	}
	return stdout.Bytes(), nil
}

// searchViaYtDlp asks yt-dlp to run the search itself with its ytsearchN:
// pseudo-URL, so the query never touches the YouTube Data API and costs no
// quota units. It returns the raw entries; filtering and mapping happen in the
// caller.
//
// An empty result set is NOT an error: yt-dlp exits 0 with "entries": [] when
// nothing matches, so callers get an empty slice and a nil error.
func searchViaYtDlp(ctx context.Context, query string, limit int) ([]model.Track, error) {
	out, err := runYtdlp(ctx,
		fmt.Sprintf("ytsearch%d:%s", limit, query),
		"--dump-single-json",
		"--flat-playlist",
	)
	if err != nil {
		return nil, err
	}

	var playlist ytdlpPlaylist
	if err := json.Unmarshal(out, &playlist); err != nil {
		return nil, fmt.Errorf("parse yt-dlp search output: %w", err)
	}

	return tracksFromEntries(playlist.Entries), nil
}

// tracksFromEntries filters and converts raw entries. Split out from
// searchViaYtDlp so the filtering rules can be tested without running yt-dlp.
func tracksFromEntries(entries []ytdlpEntry) []model.Track {
	tracks := make([]model.Track, 0, len(entries))
	for _, entry := range entries {
		if !keepVideoEntry(entry) {
			continue
		}
		tracks = append(tracks, trackFromEntry(entry))
	}
	return tracks
}

// resolveURLViaYtDlp resolves a single video URL. yt-dlp is the only backend
// that reports the duration without an API call, and it follows redirects and
// short (youtu.be) links for free.
func resolveURLViaYtDlp(ctx context.Context, videoURL string) (model.Track, error) {
	out, err := runYtdlp(ctx, videoURL, "--dump-single-json")
	if err != nil {
		return model.Track{}, err
	}

	var entry ytdlpEntry
	if err := json.Unmarshal(out, &entry); err != nil {
		return model.Track{}, fmt.Errorf("parse yt-dlp video output: %w", err)
	}
	if !keepVideoEntry(entry) {
		return model.Track{}, fmt.Errorf("not a single playable video: %s", videoURL)
	}

	track := trackFromEntry(entry)
	track.Url = videoURL // keep the URL the user pasted, it is already canonical for the streamer
	return track, nil
}

// playlistViaYtDlp expands a playlist URL into one Track per video.
//
// Flat mode is what makes this cheap: --flat-playlist reads the playlist page
// once and reports each entry's id, title and duration. Without it yt-dlp would
// fully extract all 100 videos, which is 100 network round trips for data we
// already have. The streamer later plays each track.Url on its own, so the flat
// entries are enough.
//
// No Data API fallback exists for playlists on purpose: playlistItems.list costs
// 1 unit per page of 50 plus a videos.list per 50 ids, which is exactly the quota
// the yt-dlp-first design avoids. An unreadable playlist is reported, not retried.
func playlistViaYtDlp(ctx context.Context, playlistURL string, limit int) ([]model.Track, error) {
	out, err := runYtdlp(ctx, playlistURL,
		"--dump-single-json",
		"--flat-playlist",
		fmt.Sprintf("--playlist-end=%d", limit),
	)
	if err != nil {
		return nil, mapPlaylistError(err)
	}
	return parsePlaylistOutput(out)
}

// parsePlaylistOutput unmarshals a flat playlist dump and filters the entries.
// Split from playlistViaYtDlp so the mapping is testable without the binary.
func parsePlaylistOutput(out []byte) ([]model.Track, error) {
	var playlist ytdlpPlaylist
	if err := json.Unmarshal(out, &playlist); err != nil {
		return nil, fmt.Errorf("parse yt-dlp playlist output: %w", err)
	}
	return tracksFromEntries(playlist.Entries), nil
}

// mapPlaylistError translates yt-dlp's playlist failures into ErrPlaylistUnavailable.
// The wrapped error carries stderr (see runYtdlp), and these are the three
// spellings seen in practice:
//
//	private/deleted:  ERROR: [youtube:tab] PL...: YouTube said: The playlist does not exist.
//	bad id:           ERROR: [youtube:tab] ...: Unable to download API page: HTTP Error 400
//
// Anything else (429, DNS, a missing binary) is a real fault and must stay
// visible instead of being reported to the user as "private playlist".
func mapPlaylistError(err error) error {
	message := err.Error()
	for _, marker := range []string{"does not exist", "private", "HTTP Error 400"} {
		if strings.Contains(message, marker) {
			return ErrPlaylistUnavailable
		}
	}
	return err
}

// logYtdlpFallback records why the primary backend was skipped. Logged on every
// occurrence, not once: a repeating line here is the signal that yt-dlp is
// missing, out of date, or blocked - and every one of those costs API quota.
func logYtdlpFallback(err error) {
	utils.ErrorLog.Printf("yt-dlp failed, falling back to the YouTube Data API (costs quota): %v\n", err)
}
