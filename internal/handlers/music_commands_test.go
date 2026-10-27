package handlers

import "testing"

// A search phrase must reach the provider untouched: prefixing it with
// "ytsearch:" makes providers.IsURL report true and the provider reject it.
func TestParseIdentifierKeepsSearchPhrase(t *testing.T) {
	cases := []struct{ in, want string }{
		{"daft punk", "daft punk"},
		{"never gonna give you", "never gonna give you"},
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", "https://www.youtube.com/watch?v=dQw4w9WgXcQ"},
		{"https://youtu.be/dQw4w9WgXcQ", "https://youtu.be/dQw4w9WgXcQ"},
	}
	for _, c := range cases {
		if got := parseIdentifier(c.in); got != c.want {
			t.Errorf("parseIdentifier(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
