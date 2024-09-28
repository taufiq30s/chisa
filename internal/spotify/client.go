package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const BASED_URL = "https://api.spotify.com/v1/"

type Client struct {
	httpClient *http.Client
	token      string
}

func New() (*Client, error) {
	client := &http.Client{}
	token, err := connect(client)
	if err != nil {
		return nil, err
	}

	return &Client{
		httpClient: client,
		token:      token,
	}, nil
}

func (c *Client) FindTrack(q *string) (*SearchResultDto, error) {
	// Create Search Query
	query := url.Values{}
	query.Set("q", *q)
	query.Set("type", "track")
	query.Set("limit", "10")

	// Create Request
	req, err := http.NewRequest("GET", BASED_URL+"search", strings.NewReader(query.Encode()))
	if err != nil {
		log.Fatalf("Failed to create request : %s", err)
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Fatalf("Failed to execute request : %s", err)
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body : %s", err)
		return nil, err
	}

	var res SearchResultDto
	if err := json.Unmarshal(body, &res); err != nil {
		log.Fatalf("Failed to unmarshal body : %s", err)
		return nil, err
	}
	return &res, nil
}

func (c *Client) GetTrack(id *string) (*TrackItemDto, error) {
	endpoint := BASED_URL + "tracks/" + *id
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Fatalf("Failed to create request : %s", err)
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Fatalf("Failed to execute request : %s", err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		log.Fatalf("Failed to execute request : %s", resp.Status)
		return nil, fmt.Errorf("Failed to execute request : %s", resp.Status)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body : %s", err)
		return nil, err
	}

	var res TrackItemDto
	if err := json.Unmarshal(body, &res); err != nil {
		log.Fatalf("Failed to unmarshal body : %s", err)
		return nil, err
	}
	return &res, nil
}
