package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"groupie-tracker/models"
)

// fakeArtists is the test dataset used across all API tests.
var fakeArtists = []models.Artist{
	{ID: 1, Name: "Queen", Image: "https://example.com/queen.jpg", Members: []string{"Freddie Mercury", "Brian May"}, CreationDate: 1970, FirstAlbum: "14-12-1973"},
	{ID: 2, Name: "SOJA", Image: "https://example.com/soja.jpg", Members: []string{"Jacob Hemphill"}, CreationDate: 1997, FirstAlbum: "05-09-2000"},
	{ID: 3, Name: "Mamonas Assassinas", Image: "https://example.com/mamonas.jpg", Members: []string{"Dinho"}, CreationDate: 1993, FirstAlbum: "01-06-1995"},
}

// fakeRelation is the test relation data for artist ID 1.
var fakeRelation = models.RelationData{
	ID: 1,
	DatesLocations: map[string][]string{
		"london-uk":       {"23-08-2019", "25-08-2019"},
		"new_york-usa":    {"10-12-2019"},
		"saitama-japan":   {"01-03-2020"},
	},
}

// newFakeServer starts an httptest server that serves fake JSON for
// /artists and /relation/{id} endpoints.
func newFakeServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/artists", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fakeArtists)
	})

	mux.HandleFunc("/relation", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := struct {
			Index []models.RelationData `json:"index"`
		}{
			Index: []models.RelationData{fakeRelation},
		}
		json.NewEncoder(w).Encode(response)
	})

	return httptest.NewServer(mux)
}

// setupFakeAPI starts the fake server and overrides the package-level URL
// variables so all fetches go to the local test server. Returns a cleanup
// function that restores the original state.
func setupFakeAPI(t *testing.T) func() {
	t.Helper()

	server := newFakeServer()

	// Save originals.
	origArtistsURL := artistsURL
	origRelationURL := relationURL
	origClient := httpClient

	// Point the package at the fake server.
	// These are consts, so we need to shadow them with vars visible to this package.
	// Since they are consts we can't reassign them — instead we override via the
	// unexported overrideURLs mechanism we'll set up below.
	// Actually: artistsURL and relationURL are consts, so we cannot reassign them.
	// We'll override the httpClient to redirect requests to the test server.

	// Override httpClient with a custom transport that rewrites the URL.
	httpClient = server.Client()

	// We can't change consts, but we CAN redirect via a custom RoundTripper
	// that maps the real API URLs to the test server URLs.
	httpClient.Transport = &urlRewriter{
		realArtists:  origArtistsURL,
		realRelation: origRelationURL,
		fakeBase:     server.URL,
		inner:        server.Client().Transport,
	}

	// Reset caches so each test starts clean.
	ResetCaches()

	return func() {
		server.Close()
		httpClient = origClient
		ResetCaches()
	}
}

// urlRewriter is an http.RoundTripper that replaces the real API base URL
// with the fake test server URL so the production code fetches from our local server.
type urlRewriter struct {
	realArtists  string
	realRelation string
	fakeBase     string
	inner        http.RoundTripper
}

func (u *urlRewriter) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite the request URL from real API → local test server.
	url := req.URL.String()
	if len(url) >= len(u.realArtists) && url[:len(u.realArtists)] == u.realArtists {
		req = req.Clone(req.Context())
		newURL := u.fakeBase + "/artists" + url[len(u.realArtists):]
		parsed, _ := req.URL.Parse(newURL)
		req.URL = parsed
	} else if len(url) >= len(u.realRelation) && url[:len(u.realRelation)] == u.realRelation {
		req = req.Clone(req.Context())
		newURL := u.fakeBase + "/relation" + url[len(u.realRelation):]
		parsed, _ := req.URL.Parse(newURL)
		req.URL = parsed
	}
	transport := u.inner
	if transport == nil {
		transport = http.DefaultTransport
	}
	return transport.RoundTrip(req)
}

// errorTransport is used in cache tests to ensure no network requests are made.
type errorTransport struct{}

func (errorTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("unexpected network request: should have used cache")
}

// TestGetArtists verifies that GetArtists returns the expected artists
// from our fake server (not the real API).
func TestGetArtists(t *testing.T) {
	cleanup := setupFakeAPI(t)
	defer cleanup()

	artists, err := GetArtists()
	if err != nil {
		t.Fatalf("GetArtists() returned error: %v", err)
	}

	if len(artists) != len(fakeArtists) {
		t.Errorf("GetArtists() returned %d artists, want %d", len(artists), len(fakeArtists))
	}

	// First artist should be Queen with ID 1.
	first := artists[0]
	if first.ID != 1 {
		t.Errorf("First artist ID = %d, want 1", first.ID)
	}
	if first.Name != "Queen" {
		t.Errorf("First artist Name = %q, want %q", first.Name, "Queen")
	}
}

// TestGetArtistsCache verifies that the cache is populated after the first call,
// so a second call does not hit the server again.
func TestGetArtistsCache(t *testing.T) {
	cleanup := setupFakeAPI(t)
	defer cleanup()

	// First call populates cache.
	_, err := GetArtists()
	if err != nil {
		t.Fatalf("First GetArtists() call error: %v", err)
	}

	// Disable network requests to guarantee the next call uses the cache.
	httpClient = &http.Client{Transport: errorTransport{}}

	// Second call should come from cache (even if we stopped the server it would work).
	artists, err := GetArtists()
	if err != nil {
		t.Fatalf("Second GetArtists() call error: %v", err)
	}
	if len(artists) != len(fakeArtists) {
		t.Errorf("Cached GetArtists() returned %d artists, want %d", len(artists), len(fakeArtists))
	}
}

// TestFetchRelation verifies that fetching a relation by valid ID
// returns the expected data from our fake server.
func TestFetchRelation(t *testing.T) {
	cleanup := setupFakeAPI(t)
	defer cleanup()

	relation, err := FetchRelation(1)
	if err != nil {
		t.Fatalf("FetchRelation(1) returned error: %v", err)
	}

	if relation.ID != 1 {
		t.Errorf("Relation ID = %d, want 1", relation.ID)
	}

	if len(relation.DatesLocations) == 0 {
		t.Error("FetchRelation(1) returned empty DatesLocations map")
	}

	// Check that a known location exists.
	dates, ok := relation.DatesLocations["london-uk"]
	if !ok {
		t.Error("FetchRelation(1) missing 'london-uk' key")
	}
	if len(dates) != 2 {
		t.Errorf("FetchRelation(1) london-uk dates count = %d, want 2", len(dates))
	}
}

// TestFetchRelationCache verifies the relation cache works.
func TestFetchRelationCache(t *testing.T) {
	cleanup := setupFakeAPI(t)
	defer cleanup()

	// First call fills cache.
	_, err := FetchRelation(1)
	if err != nil {
		t.Fatalf("First FetchRelation(1) error: %v", err)
	}

	// Disable network requests to guarantee the next call uses the cache.
	httpClient = &http.Client{Transport: errorTransport{}}

	// Second call should come from cache.
	relation, err := FetchRelation(1)
	if err != nil {
		t.Fatalf("Second FetchRelation(1) error: %v", err)
	}
	if relation.ID != 1 {
		t.Errorf("Cached Relation ID = %d, want 1", relation.ID)
	}
}

// TestFetchRelationInvalidID verifies that a non-existent ID returns an error
// because the fake server responds with 404.
func TestFetchRelationInvalidID(t *testing.T) {
	cleanup := setupFakeAPI(t)
	defer cleanup()

	_, err := FetchRelation(99999)
	if err == nil {
		t.Error("FetchRelation(99999) should return an error for non-existent ID")
	}
}

// TestSeedArtistCache verifies that SeedArtistCache pre-populates the cache.
func TestSeedArtistCache(t *testing.T) {
	ResetCaches()
	defer ResetCaches()

	seed := []models.Artist{
		{ID: 99, Name: "Test Band"},
	}
	SeedArtistCache(seed)

	artists, err := GetArtists()
	if err != nil {
		t.Fatalf("GetArtists() after seed returned error: %v", err)
	}
	if len(artists) != 1 || artists[0].Name != "Test Band" {
		t.Errorf("GetArtists() after seed = %v, want [{ID:99 Name:Test Band}]", artists)
	}
}

// TestSeedRelationCache verifies that SeedRelationCache pre-populates the cache.
func TestSeedRelationCache(t *testing.T) {
	ResetCaches()
	defer ResetCaches()

	data := models.RelationData{
		ID:             42,
		DatesLocations: map[string][]string{"paris-france": {"01-01-2025"}},
	}
	SeedRelationCache(42, data)

	relation, err := FetchRelation(42)
	if err != nil {
		t.Fatalf("FetchRelation(42) after seed returned error: %v", err)
	}
	if relation.ID != 42 {
		t.Errorf("Relation ID = %d, want 42", relation.ID)
	}
}

// TestCacheExpiry verifies that expired cache entries are not returned.
func TestCacheExpiry(t *testing.T) {
	cleanup := setupFakeAPI(t)
	defer cleanup()

	// Populate cache.
	_, err := GetArtists()
	if err != nil {
		t.Fatalf("GetArtists() error: %v", err)
	}

	// Force the cache to expire by backdating it, and poison the cached data
	// to ensure we can detect if it's incorrectly returned.
	artistCacheMu.Lock()
	artistCacheTime = time.Now().Add(-10 * time.Minute)
	cachedArtists[0].Name = "Poisoned Cache Data"
	artistCacheMu.Unlock()

	// Next call should re-fetch from the fake server (not fail).
	artists, err := GetArtists()
	if err != nil {
		t.Fatalf("GetArtists() after cache expiry error: %v", err)
	}

	// If the expiry check failed, it would return the poisoned cache data.
	if artists[0].Name == "Poisoned Cache Data" {
		t.Fatal("GetArtists() returned expired cache data instead of fetching fresh data")
	}

	if len(artists) != len(fakeArtists) {
		t.Errorf("GetArtists() after expiry returned %d artists, want %d", len(artists), len(fakeArtists))
	}
}
