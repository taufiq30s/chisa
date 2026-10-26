package providers

import (
	"errors"
	"path/filepath"
	"testing"
)

func duration(seconds float64) *float64 { return &seconds }

// The filter is the contract that keeps 00:00-length and unstreamable entries out
// of the picker. Each case is one measured real-world failure mode.
func TestKeepVideoEntry(t *testing.T) {
	cases := map[string]struct {
		entry ytdlpEntry
		want  bool
	}{
		"flat search entry": {
			entry: ytdlpEntry{Type: "url", IEKey: "Youtube", ID: "dQw4w9WgXcQ", Duration: duration(213)},
			want:  true,
		},
		// Full extract (a pasted watch URL) reports the same item with a different
		// spelling: _type "video" and extractor_key instead of ie_key. Regressing
		// this is invisible on the search path and breaks every direct URL.
		"full extract entry": {
			entry: ytdlpEntry{Type: "video", ExtractorKey: "Youtube", ID: "dQw4w9WgXcQ", Duration: duration(119)},
			want:  true,
		},
		"live stream without duration": {
			entry: ytdlpEntry{Type: "video", ExtractorKey: "Youtube", ID: "dQw4w9WgXcQ"},
			want:  false,
		},
		"zero duration": {
			entry: ytdlpEntry{Type: "url", IEKey: "Youtube", ID: "dQw4w9WgXcQ", Duration: duration(0)},
			want:  false,
		},
		"playlist id is not 11 chars": {
			entry: ytdlpEntry{Type: "url", IEKey: "Youtube", ID: "PLtPgu7CB4gbY9oDN3drwC3cMbJggS7dKl", Duration: duration(100)},
			want:  false,
		},
		"non-youtube extractor": {
			entry: ytdlpEntry{Type: "url", IEKey: "Vimeo", ID: "dQw4w9WgXcQ", Duration: duration(100)},
			want:  false,
		},
		"playlist container entry": {
			entry: ytdlpEntry{Type: "playlist", IEKey: "Youtube", ID: "PLtPgu7CB4g", Duration: duration(100)},
			want:  false,
		},
	}

	for name, c := range cases {
		if got := keepVideoEntry(c.entry); got != c.want {
			t.Errorf("keepVideoEntry(%s) = %v, want %v", name, got, c.want)
		}
	}
}

// A kept entry must arrive with a real length and a non-empty Url: an empty Url
// means the streamer pipes nothing into yt-dlp, and Length 0 renders as 00:00.
func TestTracksFromEntries(t *testing.T) {
	entries := []ytdlpEntry{
		{
			Type: "url", IEKey: "Youtube", ID: "dQw4w9WgXcQ", Title: "Song",
			Channel: "Artist - Topic", Duration: duration(131),
			URL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			Thumbnails: []ytdlpThumbnail{
				{URL: "https://i.ytimg.com/vi/dQw4w9WgXcQ/default.jpg", Width: 120},
				{URL: "https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg", Width: 480},
			},
		},
		{
			// No thumbnails and no URL: both must be synthesised, not left empty.
			Type: "url", IEKey: "Youtube", ID: "abcdefghijk", Title: "Bare",
			Duration: duration(59.6),
		},
		{Type: "url", IEKey: "Youtube", ID: "zzzzzzzzzzz", Title: "Live", Duration: nil},
	}

	tracks := tracksFromEntries(entries)
	if len(tracks) != 2 {
		t.Fatalf("got %d tracks, want 2 (the live stream must be dropped)", len(tracks))
	}

	first := tracks[0]
	if first.Length != 131 {
		t.Errorf("Length = %d, want 131", first.Length)
	}
	if first.Thumbnail != "https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg" {
		t.Errorf("Thumbnail = %q, want the widest available", first.Thumbnail)
	}
	if first.Artist != "Artist - Topic" {
		t.Errorf("Artist = %q, want the channel name", first.Artist)
	}

	second := tracks[1]
	if second.Length != 60 {
		t.Errorf("Length = %d, want 60 (59.6 should round)", second.Length)
	}
	if want := "https://i.ytimg.com/vi/abcdefghijk/hqdefault.jpg"; second.Thumbnail != want {
		t.Errorf("Thumbnail = %q, want %q", second.Thumbnail, want)
	}
	if want := "https://www.youtube.com/watch?v=abcdefghijk"; second.Url != want {
		t.Errorf("Url = %q, want %q", second.Url, want)
	}
}

func TestTracksFromEntriesEmpty(t *testing.T) {
	if got := tracksFromEntries(nil); len(got) != 0 {
		t.Errorf("got %d tracks from no entries, want 0", len(got))
	}
}

// A flat playlist dump is a JSON object whose entries are the same shape the
// search path already parses, so one parser covers both. Entries that the
// filter rejects (live streams) must not shorten the result silently.
func TestParsePlaylistOutput(t *testing.T) {
	out := []byte(`{
		"entries": [
			{"_type":"url","ie_key":"Youtube","id":"QCsb89zRVVE","title":"One",
			 "url":"https://www.youtube.com/watch?v=QCsb89zRVVE","duration":131},
			{"_type":"url","ie_key":"Youtube","id":"dQw4w9WgXcQ","title":"Two",
			 "url":"https://www.youtube.com/watch?v=dQw4w9WgXcQ","duration":213},
			{"_type":"url","ie_key":"Youtube","id":"zzzzzzzzzzz","title":"Live","duration":null}
		]
	}`)

	tracks, err := parsePlaylistOutput(out)
	if err != nil {
		t.Fatalf("parsePlaylistOutput: %v", err)
	}
	if len(tracks) != 2 {
		t.Fatalf("got %d tracks, want 2 (the live stream must be dropped)", len(tracks))
	}
	if tracks[0].ID != "QCsb89zRVVE" || tracks[0].Length != 131 {
		t.Errorf("first track = %s/%d, want QCsb89zRVVE/131", tracks[0].ID, tracks[0].Length)
	}
	if tracks[1].ID != "dQw4w9WgXcQ" {
		t.Errorf("second track = %s, want dQw4w9WgXcQ", tracks[1].ID)
	}
}

func TestParsePlaylistOutputInvalidJSON(t *testing.T) {
	if _, err := parsePlaylistOutput([]byte("not json")); err == nil {
		t.Fatal("want an error for unparsable yt-dlp output")
	}
}

// YouTube reports a private playlist and a nonexistent one with the same
// message, so both must map to the one user-facing error. A network fault must
// stay a real error instead of being reported as "private playlist".
func TestMapPlaylistError(t *testing.T) {
	privateErr := errors.New(`exit status 1: ERROR: [youtube:tab] PLx: YouTube said: The playlist does not exist.`)
	if got := mapPlaylistError(privateErr); !errors.Is(got, ErrPlaylistUnavailable) {
		t.Errorf("deleted/private playlist error = %v, want ErrPlaylistUnavailable", got)
	}

	badIDErr := errors.New(`exit status 1: ERROR: [youtube:tab] PLx: Unable to download API page: HTTP Error 400`)
	if got := mapPlaylistError(badIDErr); !errors.Is(got, ErrPlaylistUnavailable) {
		t.Errorf("malformed playlist id error = %v, want ErrPlaylistUnavailable", got)
	}

	networkErr := errors.New(`exit status 1: ERROR: unable to download webpage: HTTP Error 429`)
	if got := mapPlaylistError(networkErr); errors.Is(got, ErrPlaylistUnavailable) {
		t.Error("a rate-limit failure must not be reported as a private playlist")
	}
}

// YTDLP_PATH must win over the bundled copy: a deployment that ships a newer
// yt-dlp elsewhere would otherwise keep running the old one silently, and the
// failure shows up as "YouTube blocked me", not "wrong binary".
func TestFindYtdlpBinaryHonoursEnv(t *testing.T) {
	t.Setenv("YTDLP_PATH", "/opt/custom/yt-dlp")
	if got := FindYtdlpBinary(); got != "/opt/custom/yt-dlp" {
		t.Errorf("FindYtdlpBinary() = %q, want the YTDLP_PATH value", got)
	}

	// An empty value means "unset": it must fall through to the bundled binary
	// when one exists, not return "" and make exec fail with no useful message.
	t.Setenv("YTDLP_PATH", "")
	t.Chdir("../../..")
	if got := FindYtdlpBinary(); got != filepath.Join(".", "bin", "yt-dlp") {
		t.Errorf("FindYtdlpBinary() = %q, want the bundled binary", got)
	}
}
