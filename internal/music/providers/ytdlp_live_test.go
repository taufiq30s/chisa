//go:build live

// These tests drive the real yt-dlp binary and the real YouTube, so they are
// behind the "live" build tag and never run by default:
//
//	go test -tags live ./internal/music/providers/
//
// They are the only way to catch an extractor/arg regression, so run them after
// touching ytdlp.go. Everything else in this package is offline.
package providers

import (
	"context"
	"errors"
	"testing"
)

// A cold search must return playable tracks: a real Length (an empty one renders
// as 00:00) and a non-empty Url (the streamer pipes it into yt-dlp).
func TestYtDlpSearchLive(t *testing.T) {
	t.Chdir("../../..")

	tracks, err := searchViaYtDlp(context.Background(), "daft punk", 5)
	if err != nil {
		t.Fatalf("searchViaYtDlp: %v", err)
	}
	if len(tracks) == 0 {
		t.Fatal("no tracks; either yt-dlp is missing or the filter dropped everything")
	}
	for _, track := range tracks {
		if track.Length <= 0 {
			t.Errorf("%q: Length = %d, want > 0", track.Title, track.Length)
		}
		if track.Url == "" {
			t.Errorf("%q: empty Url, the streamer cannot play it", track.Title)
		}
		if track.Thumbnail == "" {
			t.Errorf("%q: empty Thumbnail", track.Title)
		}
	}

	// A query with no matching video must be an empty slice, not an error.
	empty, err := searchViaYtDlp(context.Background(), "zzzqqqxxxvvvnoresult12345", 5)
	if err != nil {
		t.Fatalf("empty search returned an error: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("got %d tracks for a nonsense query, want 0", len(empty))
	}
}

// Search uses --flat-playlist, direct URLs use full extraction, and the two
// report the item with different field names - the exact drift that made direct
// URLs fail while search kept working.
func TestResolveURLViaYtDlpLive(t *testing.T) {
	t.Chdir("../../..")

	track, err := resolveURLViaYtDlp(context.Background(),
		"https://www.youtube.com/watch?v=QCsb89zRVVE")
	if err != nil {
		t.Fatalf("resolveURLViaYtDlp: %v", err)
	}
	if track.Length != 119 {
		t.Errorf("Length = %d, want 119", track.Length)
	}
	if track.ID != "QCsb89zRVVE" {
		t.Errorf("ID = %q, want QCsb89zRVVE", track.ID)
	}
	if track.Thumbnail == "" {
		t.Error("empty Thumbnail")
	}
}

// Proves the real binary honours --playlist-end and that a private playlist maps
// to ErrPlaylistUnavailable.
func TestPlaylistViaYtDlpLive(t *testing.T) {
	t.Chdir("../../..")

	const publicPlaylist = "https://www.youtube.com/playlist?list=PLLUhuJBynpzWx8t9aJ417z9NrwyzF0i3R"
	tracks, err := playlistViaYtDlp(context.Background(), publicPlaylist, 3)
	if err != nil {
		t.Fatalf("playlistViaYtDlp: %v", err)
	}
	if len(tracks) != 3 {
		t.Fatalf("got %d tracks, want 3 (--playlist-end must cap the dump)", len(tracks))
	}
	for _, track := range tracks {
		if track.Length <= 0 {
			t.Errorf("%q: Length = %d, want > 0", track.Title, track.Length)
		}
		if track.Url == "" {
			t.Errorf("%q: empty Url, the streamer cannot play it", track.Title)
		}
	}

	// YouTube hides whether a playlist is private or missing behind the same
	// message, so both must surface as the same friendly error.
	const privatePlaylist = "https://www.youtube.com/playlist?list=PLtPgu7CB4gbY9oDN3drwC3cMbJggS7dKl"
	if _, err := playlistViaYtDlp(context.Background(), privatePlaylist, 3); !errors.Is(err, ErrPlaylistUnavailable) {
		t.Errorf("private playlist error = %v, want ErrPlaylistUnavailable", err)
	}
}
