package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

const (
	googleUserInfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"
	minAliasLen       = 4
	maxAliasLen       = 64
)

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func fetchGoogleUserInfo(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (googleUserInfo, error) {
	client := cfg.Client(ctx, token)

	resp, err := client.Get(googleUserInfoURL)
	if err != nil {
		return googleUserInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return googleUserInfo{}, fmt.Errorf("google userinfo: unexpected status %d", resp.StatusCode)
	}

	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return googleUserInfo{}, err
	}

	return info, nil
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
