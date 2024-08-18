package spotify

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/taufiq30s/chisa/utils"
)

const authUrl = "https://accounts.spotify.com/api/token"

type spotifyAuthResponse struct {
	AccessToken string `json:"access_token"`
}

func connect(client *http.Client) (string, error) {
	clientId, err := utils.GetEnv("SPOTIFY_CLIENT_ID")
	if err != nil {
		log.Fatalf("Failed to load .env : %s", err)
		return "", err
	}

	clientSecret, err := utils.GetEnv("SPOTIFY_CLIENT_SECRET")
	if err != nil {
		log.Fatalf("Failed to load .env : %s", err)
		return "", err
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", authUrl, strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatalf("Failed to create request : %s", err)
		return "", err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientId, clientSecret)

	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to execute request : %s", err)
		return "", err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Failed to read response body : %s", err)
		return "", err
	}
	var resp spotifyAuthResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		log.Fatalf("Failed to unmarshal body : %s", err)
		return "", err
	}
	return resp.AccessToken, nil
}
