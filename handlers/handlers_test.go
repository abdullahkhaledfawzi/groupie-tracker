package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"groupie-tracker/models"
)

// TestMain sets up the test environment by changing to the project root
// so that template files can be found by ParseGlob("templates/*.html").
func TestMain(m *testing.M) {
	// Walk up from handlers/ to project root.
	projectRoot, _ := filepath.Abs("..")
	os.Chdir(projectRoot)
	InitTemplates()
	os.Exit(m.Run())
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
	if !strings.Contains(body, "Groupie Tracker") {
		t.Error("HomeHandler response does not contain 'Groupie Tracker'")
	}
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

// TestArtistHandlerNotFound verifies out-of-range ID returns 404.
func TestArtistHandlerNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/artist?id=9999", nil)
	rec := httptest.NewRecorder()

	ArtistHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("ArtistHandler GET /artist?id=9999 status = %d, want %d", rec.Code, http.StatusNotFound)
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

// TestSearchHandlerEmpty verifies empty query returns all artists.
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

	if len(results) != 52 {
		t.Errorf("SearchHandler empty query returned %d artists, want 52", len(results))
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
