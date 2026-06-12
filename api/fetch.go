// Package api handles all communication with the external Groupie Trackers API.
// It provides functions to fetch and decode JSON data from the API endpoints.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"groupie-tracker/models"
)

// API endpoint constants.
const (
	baseURL     = "https://groupietrackers.herokuapp.com/api"
	artistsURL  = baseURL + "/artists"
	relationURL = baseURL + "/relation"
)

// httpClient is a shared HTTP client with a 10-second timeout
// to prevent hanging connections if the external API is unresponsive.
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// In-memory cache for artists data to avoid redundant API calls.
var (
	cachedArtists []models.Artist
	cacheMu       sync.Mutex
	cacheTime     time.Time
	cacheDuration = 5 * time.Minute
)

// GetArtists returns all artists, using a cached version if available and fresh.
// This avoids hitting the external API on every request (e.g., search, home page).
func GetArtists() ([]models.Artist, error) {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	if cachedArtists != nil && time.Since(cacheTime) < cacheDuration {
		return cachedArtists, nil
	}

	artists, err := fetchArtists()
	if err != nil {
		return nil, err
	}

	cachedArtists = artists
	cacheTime = time.Now()
	return artists, nil
}

// fetchArtists retrieves all artists from the API.
func fetchArtists() ([]models.Artist, error) {
	var artists []models.Artist
	if err := fetchJSON(artistsURL, &artists); err != nil {
		return nil, fmt.Errorf("fetchArtists: %w", err)
	}

	// Override image for Mamonas Assassinas (id=21) without an extra loop,
	// by checking the 20th index directly.
	if len(artists) >= 21 && artists[20].ID == 21 {
		artists[20].Image = "/static/Mamonas Assassinas.jpg"
	}

	return artists, nil
}

// FetchRelation retrieves the relation data (location→dates mapping)
// for a specific artist by their ID.
func FetchRelation(id int) (models.RelationData, error) {
	url := fmt.Sprintf("%s/%d", relationURL, id)
	var relation models.RelationData
	if err := fetchJSON(url, &relation); err != nil {
		return relation, fmt.Errorf("FetchRelation(%d): %w", id, err)
	}
	return relation, nil
}

// fetchJSON is a DRY helper that performs an HTTP GET request
// and decodes the JSON response body into the provided target.
func fetchJSON(url string, target any) error {
	resp, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP GET %s: unexpected status %d", url, resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("JSON decode from %s: %w", url, err)
	}

	return nil
}
