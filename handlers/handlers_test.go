package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"groupie-tracker/api"
	"groupie-tracker/models"
)

// testArtists is the fake dataset used across all handler tests.
var testArtists = []models.Artist{
	{ID: 1, Name: "Queen", Image: "/static/queen.jpg", Members: []string{"Freddie Mercury", "Brian May"}, CreationDate: 1970, FirstAlbum: "14-12-1973"},
	{ID: 2, Name: "SOJA", Image: "/static/soja.jpg", Members: []string{"Jacob Hemphill"}, CreationDate: 1997, FirstAlbum: "05-09-2000"},
	{ID: 3, Name: "Pink Floyd", Image: "/static/pinkfloyd.jpg", Members: []string{"Roger Waters", "David Gilmour"}, CreationDate: 1965, FirstAlbum: "05-08-1967"},
}

// testRelation is the fake relation data for artist ID 1.
var testRelation = models.RelationData{
	ID: 1,
	DatesLocations: map[string][]string{
		"london-uk":    {"23-08-2019"},
		"new_york-usa": {"10-12-2019"},
	},
}

// TestMain sets up the test environment: changes to the project root so
// templates can be found, seeds the API caches with fake data, and runs tests.
func TestMain(m *testing.M) {
	// Walk up from handlers/ to project root.
	projectRoot, _ := filepath.Abs("..")
	os.Chdir(projectRoot)
	InitTemplates()

	// Seed API caches so handlers never call the real external API.
	api.SeedArtistCache(testArtists)
	api.SeedRelationCache(1, testRelation)

	code := m.Run()

	// Clean up caches after all tests.
	api.ResetCaches()
	os.Exit(code)
}

// TestHomeHandler verifies the home page returns 200 and contains expected HTML.
func TestHomeHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	HomeHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("HomeHandler GET / status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if len(body) == 0 {
		t.Error("HomeHandler response body is empty")
	}
	// Should contain the site name somewhere in the HTML.
	if !strings.Contains(body, "Groupie Tracker") {
		t.Error("HomeHandler response does not contain 'Groupie Tracker'")
	}
	// Should contain our seeded artist name.
	if !strings.Contains(body, "Queen") {
		t.Error("HomeHandler response does not contain 'Queen'")
	}
}

// TestHomeHandler404 verifies that unknown paths return 404.
func TestHomeHandler404(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()

	HomeHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("HomeHandler GET /nonexistent status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// TestHomeHandlerMethodNotAllowed verifies POST returns 405.
func TestHomeHandlerMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	HomeHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("HomeHandler POST / status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

// TestArtistHandlerValid verifies a valid artist ID returns 200.
func TestArtistHandlerValid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/artist?id=1", nil)
	rec := httptest.NewRecorder()

	ArtistHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("ArtistHandler GET /artist?id=1 status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Queen") {
		t.Error("ArtistHandler response for id=1 does not contain 'Queen'")
	}
}

// TestArtistHandlerMissingID verifies missing ID returns 400.
func TestArtistHandlerMissingID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/artist", nil)
	rec := httptest.NewRecorder()

	ArtistHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("ArtistHandler GET /artist status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// TestArtistHandlerInvalidID verifies non-numeric ID returns 400.
func TestArtistHandlerInvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/artist?id=abc", nil)
	rec := httptest.NewRecorder()

	ArtistHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("ArtistHandler GET /artist?id=abc status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// TestArtistHandlerNegativeID verifies that a negative ID returns 404
// immediately via the lower-bound check (no need to iterate all artists).
func TestArtistHandlerNegativeID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/artist?id=-5", nil)
	rec := httptest.NewRecorder()

	ArtistHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("ArtistHandler GET /artist?id=-5 status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// TestArtistHandlerZeroID verifies that ID=0 returns 404.
func TestArtistHandlerZeroID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/artist?id=0", nil)
	rec := httptest.NewRecorder()

	ArtistHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("ArtistHandler GET /artist?id=0 status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// TestArtistHandlerNotFound verifies out-of-range ID returns 404.
func TestArtistHandlerNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/artist?id=9999", nil)
	rec := httptest.NewRecorder()

	ArtistHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("ArtistHandler GET /artist?id=9999 status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// TestArtistHandlerMethodNotAllowed verifies POST to artist page returns 405.
func TestArtistHandlerMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/artist?id=1", nil)
	rec := httptest.NewRecorder()

	ArtistHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("ArtistHandler POST /artist?id=1 status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

// TestSearchHandler verifies the search endpoint returns matching artists as JSON.
func TestSearchHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/search?q=queen", nil)
	rec := httptest.NewRecorder()

	SearchHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("SearchHandler GET /search?q=queen status = %d, want %d", rec.Code, http.StatusOK)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("SearchHandler Content-Type = %q, want application/json", contentType)
	}

	var results []models.Artist
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("SearchHandler response is not valid JSON: %v", err)
	}

	if len(results) == 0 {
		t.Error("SearchHandler for 'queen' returned no results")
	}

	found := false
	for _, a := range results {
		if a.Name == "Queen" {
			found = true
			break
		}
	}
	if !found {
		t.Error("SearchHandler for 'queen' did not return Queen")
	}
}

// TestSearchHandlerNoMatch verifies that a query with no matches returns an empty array.
func TestSearchHandlerNoMatch(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/search?q=zzzznonexistent", nil)
	rec := httptest.NewRecorder()

	SearchHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("SearchHandler status = %d, want %d", rec.Code, http.StatusOK)
	}

	var results []models.Artist
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("SearchHandler response is not valid JSON: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("SearchHandler for 'zzzznonexistent' returned %d results, want 0", len(results))
	}
}

// TestSearchHandlerEmpty verifies empty query returns all seeded artists.
func TestSearchHandlerEmpty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/search?q=", nil)
	rec := httptest.NewRecorder()

	SearchHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("SearchHandler GET /search?q= status = %d, want %d", rec.Code, http.StatusOK)
	}

	var results []models.Artist
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("SearchHandler response is not valid JSON: %v", err)
	}

	if len(results) != len(testArtists) {
		t.Errorf("SearchHandler empty query returned %d artists, want %d", len(results), len(testArtists))
	}
}

// TestSearchHandlerMethodNotAllowed verifies POST to search returns 405.
func TestSearchHandlerMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/search?q=test", nil)
	rec := httptest.NewRecorder()

	SearchHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("SearchHandler POST status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

// TestFormatLocation verifies the location string formatting function.
func TestFormatLocation(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"north_carolina-usa", "North Carolina, USA"},
		{"georgia-usa", "Georgia, USA"},
		{"saitama-japan", "Saitama, Japan"},
		{"london-uk", "London, UK"},
		{"abu_dhabi-united_arab_emirates", "Abu Dhabi, United Arab Emirates"},
		{"penrose-new_zealand", "Penrose, New Zealand"},
	}

	for _, tt := range tests {
		got := FormatLocation(tt.input)
		if got != tt.want {
			t.Errorf("FormatLocation(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}


