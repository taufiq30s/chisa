package providers

import "testing"

// isPlaylistURL decides whether a pasted URL becomes a playlist import (many
// tracks, one summary card) or a single track. Getting "watch?v=X&list=Y" wrong
// would queue the whole playlist when the user asked for one video.
func TestIsPlaylistURL(t *testing.T) {
	cases := map[string]bool{
		"https://www.youtube.com/playlist?list=PLLUhuJBynpzWx8t9aJ417z9NrwyzF0i3R": true,
		"https://www.youtube.com/playlist?list=PLabc":                              true,
		"https://m.youtube.com/playlist?list=PLabc":                                true,
		"https://www.youtube.com/videoseries?list=PLabc":                           true,
		"https://www.youtube.com/playlist?list=":                                   false,
		// v= wins: the user asked for the video, not the list it belongs to.
		"https://www.youtube.com/watch?v=QCsb89zRVVE&list=PLabc": false,
		"https://www.youtube.com/watch?v=QCsb89zRVVE":            false,
		"https://youtu.be/QCsb89zRVVE":                           false,
		"https://www.youtube.com/watch?list=PLabc":               false,
		"https://example.com/playlist?list=PLabc":                false,
		"never gonna give you up":                                false,
		"":                                                       false,
	}

	for query, want := range cases {
		if got := isPlaylistURL(query); got != want {
			t.Errorf("isPlaylistURL(%q) = %v, want %v", query, got, want)
		}
	}
}

// isPlaylistURL runs before any network call, so garbage must be rejected
// without panicking on a bad URL.
func TestIsPlaylistURLMalformed(t *testing.T) {
	for _, query := range []string{"://", "youtube.com/playlist?list=PLab", "%%%"} {
		if isPlaylistURL(query) {
			t.Errorf("isPlaylistURL(%q) = true, want false", query)
		}
	}
}
