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
// cacheDuration is shared by both the artists and relation caches.
var cacheDuration = 5 * time.Minute

// In-memory cache for the artists list.
var (
	cachedArtists  []models.Artist
	artistCacheMu  sync.RWMutex
	artistCacheTime time.Time
)

// In-memory cache for all relation data, keyed by artist ID.
var (
	relationCache     map[int]models.RelationData
	relationCacheMu   sync.RWMutex
	relationCacheTime time.Time
)

// GetArtists returns all artists, using a cached version if available and fresh.
// This avoids hitting the external API on every request (e.g., search, home page).
// Uses RLock first so multiple goroutines can read concurrently.
func GetArtists() ([]models.Artist, error) {
	// Fast path: read lock for cache hit.
	artistCacheMu.RLock()
	if cachedArtists != nil && time.Since(artistCacheTime) < cacheDuration {
		defer artistCacheMu.RUnlock()
		return cachedArtists, nil
	}
	artistCacheMu.RUnlock()

	// Slow path: write lock to refresh the cache.
	artistCacheMu.Lock()
	defer artistCacheMu.Unlock()

	// Double-check after acquiring write lock (another goroutine may have refreshed).
	if cachedArtists != nil && time.Since(artistCacheTime) < cacheDuration {
		return cachedArtists, nil
	}

	artists, err := fetchArtists()
	if err != nil {
		return nil, err
	}

	cachedArtists = artists
	artistCacheTime = time.Now()
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

// getRelations returns all relations, using a cached version if available and fresh.
func getRelations() (map[int]models.RelationData, error) {
	relationCacheMu.RLock()
	if relationCache != nil && time.Since(relationCacheTime) < cacheDuration {
		defer relationCacheMu.RUnlock()
		return relationCache, nil
	}
	relationCacheMu.RUnlock()

	relationCacheMu.Lock()
	defer relationCacheMu.Unlock()

	if relationCache != nil && time.Since(relationCacheTime) < cacheDuration {
		return relationCache, nil
	}

	var response struct {
		Index []models.RelationData `json:"index"`
	}
	if err := fetchJSON(relationURL, &response); err != nil {
		return nil, fmt.Errorf("fetchAllRelations: %w", err)
	}

	relationsMap := make(map[int]models.RelationData)
	for _, rel := range response.Index {
		relationsMap[rel.ID] = rel
	}

	relationCache = relationsMap
	relationCacheTime = time.Now()
	return relationsMap, nil
}

// FetchRelation retrieves the relation data (location→dates mapping)
// for a specific artist by their ID, using a 5-minute global cache.
func FetchRelation(id int) (models.RelationData, error) {
	relations, err := getRelations()
	if err != nil {
		return models.RelationData{}, err
	}

	rel, ok := relations[id]
	if !ok {
		return models.RelationData{}, fmt.Errorf("FetchRelation(%d): not found", id)
	}
	return rel, nil
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

// SeedArtistCache pre-populates the artist cache with the given data.
// Intended for use by tests so they can run without hitting the real API.
func SeedArtistCache(artists []models.Artist) {
	artistCacheMu.Lock()
	defer artistCacheMu.Unlock()
	cachedArtists = artists
	artistCacheTime = time.Now()
}

// SeedRelationCache pre-populates the relation cache for a given artist ID.
// Intended for use by tests so they can run without hitting the real API.
func SeedRelationCache(id int, data models.RelationData) {
	relationCacheMu.Lock()
	defer relationCacheMu.Unlock()
	if relationCache == nil {
		relationCache = make(map[int]models.RelationData)
	}
	relationCache[id] = data
	relationCacheTime = time.Now()
}

// ResetCaches clears both the artist and relation caches.
// Intended for use by tests to ensure a clean state between test runs.
func ResetCaches() {
	artistCacheMu.Lock()
	cachedArtists = nil
	artistCacheTime = time.Time{}
	artistCacheMu.Unlock()

	relationCacheMu.Lock()
	relationCache = nil
	relationCacheTime = time.Time{}
	relationCacheMu.Unlock()
}
