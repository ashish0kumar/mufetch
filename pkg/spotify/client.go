package spotify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client represents a Spotify API client with authentication
type Client struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
	TokenExpiry  time.Time
}

// TokenResponse represents the OAuth token response from Spotify
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// SearchResponse represents the combined search results from Spotify API
type SearchResponse struct {
	Tracks  TracksResponse  `json:"tracks"`
	Albums  AlbumsResponse  `json:"albums"`
	Artists ArtistsResponse `json:"artists"`
}

// TracksResponse represents the tracks section of search results
type TracksResponse struct {
	Items []Track `json:"items"`
}

// AlbumsResponse represents the albums section of search results
type AlbumsResponse struct {
	Items []Album `json:"items"`
}

// ArtistsResponse represents the artists section of search results
type ArtistsResponse struct {
	Items []Artist `json:"items"`
}

// Track represents a Spotify track with all metadata
type Track struct {
	ID               string       `json:"id"`
	Name             string       `json:"name"`
	Artists          []Artist     `json:"artists"`
	Album            Album        `json:"album"`
	Duration         int          `json:"duration_ms"`
	Popularity       int          `json:"popularity"`
	TrackNumber      int          `json:"track_number"`
	DiscNumber       int          `json:"disc_number"`
	Explicit         bool         `json:"explicit"`
	PreviewURL       string       `json:"preview_url"`
	ExternalURL      ExternalURL  `json:"external_urls"`
	AvailableMarkets []string     `json:"available_markets"`
	Restrictions     Restrictions `json:"restrictions"`
}

// Album represents a Spotify album with all metadata
type Album struct {
	ID                   string       `json:"id"`
	Name                 string       `json:"name"`
	Artists              []Artist     `json:"artists"`
	Images               []Image      `json:"images"`
	ReleaseDate          string       `json:"release_date"`
	ReleaseDatePrecision string       `json:"release_date_precision"`
	TotalTracks          int          `json:"total_tracks"`
	Genres               []string     `json:"genres"`
	Popularity           int          `json:"popularity"`
	AlbumType            string       `json:"album_type"`
	AlbumGroup           string       `json:"album_group"`
	Label                string       `json:"label"`
	Copyrights           []Copyright  `json:"copyrights"`
	ExternalURL          ExternalURL  `json:"external_urls"`
	AvailableMarkets     []string     `json:"available_markets"`
	Restrictions         Restrictions `json:"restrictions"`
	Tracks               TracksPage   `json:"tracks"`
}

// Artist represents a Spotify artist with all metadata
type Artist struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Images      []Image     `json:"images"`
	Genres      []string    `json:"genres"`
	Popularity  int         `json:"popularity"`
	Followers   Followers   `json:"followers"`
	ExternalURL ExternalURL `json:"external_urls"`
	Type        string      `json:"type"`
}

// Image represents cover art or artist image metadata
type Image struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

// Followers represents artist follower count
type Followers struct {
	Total int `json:"total"`
}

// ExternalURL represents external URLs for Spotify entities
type ExternalURL struct {
	Spotify string `json:"spotify"`
}

// Copyright represents album copyright information
type Copyright struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

// Restrictions represents content restrictions
type Restrictions struct {
	Reason string `json:"reason"`
}

// TracksPage represents a paginated list of tracks
type TracksPage struct {
	Items []Track `json:"items"`
	Total int     `json:"total"`
}

// TopTracksResponse represents artist's top tracks response
type TopTracksResponse struct {
	Tracks []Track `json:"tracks"`
}

// RelatedArtistsResponse represents related artists response
type RelatedArtistsResponse struct {
	Artists []Artist `json:"artists"`
}

// ArtistAlbumsResponse represents artist's albums response
type ArtistAlbumsResponse struct {
	Items []Album `json:"items"`
	Total int     `json:"total"`
}

// NewClient creates a new Spotify API client with credentials
func NewClient(clientID, clientSecret string) *Client {
	return &Client{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
}

// authenticate obtains or refreshes the access token for API calls
func (c *Client) authenticate() error {
	if time.Now().Before(c.TokenExpiry) {
		return nil // Token still valid
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	// Create a new HTTP request for token endpoint
	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	// Encode credentials for Basic authentication
	auth := base64.StdEncoding.EncodeToString([]byte(c.ClientID + ":" + c.ClientSecret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed: %s", resp.Status)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	// Store token and expiry time
	c.AccessToken = tokenResp.AccessToken
	c.TokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

// Search performs a search query for tracks, albums, or artists
func (c *Client) Search(query, searchType string) (*SearchResponse, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("type", searchType)
	params.Set("limit", "1")

	reqURL := "https://api.spotify.com/v1/search?" + params.Encode()

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed: %s - %s", resp.Status, string(body))
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	return &searchResp, nil
}

// GetAlbum retrieves detailed album information by ID
func (c *Client) GetAlbum(albumID string) (*Album, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("https://api.spotify.com/v1/albums/%s", albumID)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get album: %s", resp.Status)
	}

	var album Album
	if err := json.NewDecoder(resp.Body).Decode(&album); err != nil {
		return nil, err
	}

	return &album, nil
}

// GetArtist retrieves detailed artist information by ID
func (c *Client) GetArtist(artistID string) (*Artist, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("https://api.spotify.com/v1/artists/%s", artistID)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get artist: %s", resp.Status)
	}

	var artist Artist
	if err := json.NewDecoder(resp.Body).Decode(&artist); err != nil {
		return nil, err
	}

	return &artist, nil
}

// GetArtistTopTracks retrieves an artist's most popular tracks
func (c *Client) GetArtistTopTracks(artistID string) (*TopTracksResponse, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("https://api.spotify.com/v1/artists/%s/top-tracks?market=US", artistID)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get top tracks: %s", resp.Status)
	}

	var topTracks TopTracksResponse
	if err := json.NewDecoder(resp.Body).Decode(&topTracks); err != nil {
		return nil, err
	}

	return &topTracks, nil
}

// GetArtistAlbums retrieves an artist's albums by type (album, single, etc.)
func (c *Client) GetArtistAlbums(artistID string, includeGroups string) (*ArtistAlbumsResponse, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("include_groups", includeGroups)
	params.Set("limit", "50")
	params.Set("market", "US")

	reqURL := fmt.Sprintf("https://api.spotify.com/v1/artists/%s/albums?%s", artistID, params.Encode())

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get artist albums: %s", resp.Status)
	}

	var albums ArtistAlbumsResponse
	if err := json.NewDecoder(resp.Body).Decode(&albums); err != nil {
		return nil, err
	}

	return &albums, nil
}
