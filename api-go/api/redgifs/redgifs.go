package redgifs

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"kannonfoundry/api-go/api"
	"net/http"
	"strings"
	"time"
)

type RedGifsClient struct {
	authToken   string
	tokenExpiry int64
}
type loginResponse struct {
	Token string `json:"token"`
}

func NewClient() *RedGifsClient {
	return &RedGifsClient{}
}

var (
	baseUrl = "https://api.redgifs.com/v2"
)

func (client *RedGifsClient) login() error {
	client.authToken = "" // reset any existing token
	resp, err := http.Get(baseUrl + "/auth/temporary")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to login: %s", resp.Status)
	}

	var loginResp loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return err
	}

	client.authToken = loginResp.Token

	// Decode JWT token to extract exp
	parts := strings.Split(loginResp.Token, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid JWT token format")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("failed to decode JWT payload: %w", err)
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return fmt.Errorf("failed to parse JWT claims: %w", err)
	}
	client.tokenExpiry = claims.Exp

	return nil
}

func (c *RedGifsClient) IsTokenExpired() bool {
	if c.tokenExpiry == 0 {
		return true // treat unset expiry as expired
	}
	return time.Now().Unix() >= c.tokenExpiry
}

type UrlResponse struct {
	Hd string `json:"hd"`
	Sd string `json:"sd"`
}
type GifResponse struct {
	Urls     UrlResponse  `json:"urls"`
	Id       string       `json:"id"`
	CTA      *CtaResponse `json:"cta"`
	Username string       `json:"userName"`
}
type CtaResponse struct {
	ShowReason string `json:"showReason"`
}
type GifsResponse struct {
	Gifs []GifResponse `json:"gifs"`
}

func (c *RedGifsClient) FormatAndModifySearch(tags []string, authorID int64) (searchTerm string, err error) {
	return strings.Join(tags, "|"), nil
}

func (c *RedGifsClient) Search(tags []string, count int, page int) (files []api.FileToSend, err error) {
	if c.IsTokenExpired() {
		if err := c.login(); err != nil {
			return nil, fmt.Errorf("failed to login: %w", err)
		}
	}
	req, err := http.NewRequest("GET", baseUrl+"/gifs/search?type=g&order=latest&page="+fmt.Sprint(page)+"&count="+fmt.Sprint(count)+"&tags="+strings.Join(tags, "|"), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search request failed: %s", resp.Status)
	}
	var searchResp GifsResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}
	var results []api.FileToSend
	for _, gif := range searchResp.Gifs {
		if gif.CTA != nil {
			continue //skip adverts
		}
		toFileToSend := toFileToSend(gif)
		if toFileToSend != nil {
			results = append(results, *toFileToSend)
		}
	}
	return results, nil
}

func (c *RedGifsClient) SearchByUser(username string, count int, page int) (files []api.FileToSend, err error) {
	if c.IsTokenExpired() {
		if err := c.login(); err != nil {
			return nil, fmt.Errorf("failed to login: %w", err)
		}
	}
	req, err := http.NewRequest("GET", baseUrl+"/users/"+username+"/search?type=g&order=new&page="+fmt.Sprint(page)+"&count="+fmt.Sprint(count), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search request failed: %s", resp.Status)
	}
	var searchResp GifsResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}
	var results []api.FileToSend
	for _, gif := range searchResp.Gifs {
		if gif.CTA != nil {
			continue //skip adverts
		}
		toFileToSend := toFileToSend(gif)
		if toFileToSend != nil {
			results = append(results, *toFileToSend)
		}
	}
	return results, nil
}

func FormatFileUrls(files []api.FileToSend) {
	for i, file := range files {
		files[i].URL = strings.Replace(file.URL, "https://media.redgifs.com/", "/rgp/", 1)
	}
}

func toFileToSend(gif GifResponse) *api.FileToSend {
	url := gif.Urls.Sd
	if url == "" {
		url = gif.Urls.Hd
	}
	if url == "" {
		return nil
	}
	return &api.FileToSend{
		Name:     "redgif_" + gif.Id,
		URL:      url,
		Username: gif.Username,
	}
}
