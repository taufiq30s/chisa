package spotify

type SearchResultDto struct {
	Href     string         `json:"href"`
	Limit    uint           `json:"limit"`
	Total    uint           `json:"total"`
	Offset   uint           `json:"offset"`
	Previous string         `json:"previous,omitempty"`
	Next     string         `json:"next,omitempty"`
	Items    []TrackItemDto `json:"items"`
}

type TrackItemDto struct {
	Album            AlbumDto        `json:"album"`
	Artists          []ArtistDto     `json:"artists"`
	AvailableMarkets []string        `json:"available_markets"`
	DiscNumber       int             `json:"disc_number"`
	DurationMs       int             `json:"duration_ms"`
	Explicit         bool            `json:"explicit"`
	ExternalIds      ExternalIdsDto  `json:"external_ids"`
	ExternalUrls     ExternalUrlsDto `json:"external_urls"`
	Href             string          `json:"href"`
	Id               string          `json:"id"`
	IsPlayable       bool            `json:"is_playable"`
	LinkedFrom       interface{}     `json:"linked_from"`
	Restrictions     RestrictionsDto `json:"restrictions"`
	Name             string          `json:"name"`
	Popularity       int             `json:"popularity"`
	PreviewUrl       string          `json:"preview_url"`
	TrackNumber      int             `json:"track_number"`
	Type             string          `json:"type"`
	Uri              string          `json:"uri"`
	IsLocal          bool            `json:"is_local"`
}

type AlbumDto struct {
	AlbumType            string          `json:"album_type"`
	TotalTracks          int             `json:"total_tracks"`
	AvailableMarkets     []string        `json:"available_markets"`
	ExternalUrls         ExternalUrlsDto `json:"external_urls"`
	Href                 string          `json:"href"`
	Id                   string          `json:"id"`
	Images               []ImageDto      `json:"images"`
	Name                 string          `json:"name"`
	ReleaseDate          string          `json:"release_date"`
	ReleaseDatePrecision string          `json:"release_date_precision"`
	Restrictions         RestrictionsDto `json:"restrictions"`
	Type                 string          `json:"type"`
	Uri                  string          `json:"uri"`
	Artists              []ArtistDto     `json:"artists"`
}

type ArtistDto struct {
	ExternalUrls ExternalUrlsDto `json:"external_urls"`
	Href         string          `json:"href"`
	Id           string          `json:"id"`
	Name         string          `json:"name"`
	Type         string          `json:"type"`
	Uri          string          `json:"uri"`
}

type ExternalIdsDto struct {
	Isrc string `json:"isrc"`
	Ean  string `json:"ean"`
	Upc  string `json:"upc"`
}

type ExternalUrlsDto struct {
	Spotify string `json:"spotify"`
}

type ImageDto struct {
	Url    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type RestrictionsDto struct {
	Reason string `json:"reason"`
}
