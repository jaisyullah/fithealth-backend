package oauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type TokenManager struct {
	token     string
	expiresAt time.Time
	mu        sync.Mutex

	tokenURL string
	clientID string
	secret   string
	timeout  time.Duration
}

type tokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func NewTokenManager(tokenURL, clientID, secret string, timeout time.Duration) *TokenManager {
	return &TokenManager{
		tokenURL: tokenURL,
		clientID: clientID,
		secret:   secret,
		timeout:  timeout,
	}
}

func (m *TokenManager) GetToken() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.token != "" && time.Now().Before(m.expiresAt.Add(-30*time.Second)) {
		return m.token, nil
	}
	v := url.Values{}
	v.Set("grant_type", "client_credentials")
	v.Set("client_id", m.clientID)
	v.Set("client_secret", m.secret)
	req, err := http.NewRequest("POST", m.tokenURL, bytes.NewBufferString(v.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{Timeout: m.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("token request failed status %d", resp.StatusCode)
	}
	var tr tokenResp
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", err
	}
	m.token = tr.AccessToken
	m.expiresAt = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	return m.token, nil
}
