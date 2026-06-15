// Package handlers implements the HTTP request handlers for the web application.
// It processes incoming requests, fetches data from the API layer,
// and renders HTML templates or JSON responses.
package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"groupie-tracker/api"
	"groupie-tracker/models"
)

// templates holds all parsed HTML templates, initialized once at startup.
var templates *template.Template

// InitTemplates parses all HTML templates from the templates directory.
// Must be called once before handling any requests.
// Panics if templates are missing or invalid (acceptable at startup).
func InitTemplates() {
	funcMap := template.FuncMap{
		"formatLocation": FormatLocation,
	}
	templates = template.Must(
		template.New("").Funcs(funcMap).ParseGlob("templates/*.html"),
	)
}

// HomeHandler serves the home page displaying a grid of all artist cards.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Only the exact "/" path should be handled here.
	// All other unregistered paths must return 404.
	if r.URL.Path != "/" {
		renderError(w, http.StatusNotFound, "Page Not Found")
		return
	}

	if r.Method != http.MethodGet {
		renderError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	artists, err := api.GetArtists()
	if err != nil {
		log.Printf("HomeHandler: failed to fetch artists: %v", err)
		renderError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, "index.html", artists); err != nil {
		log.Printf("HomeHandler: template execution error: %v", err)
	}
}

// ArtistHandler serves the detail page for a single artist.
// Expects a query parameter "id" (e.g., /artist?id=1).
func ArtistHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		renderError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// Validate that an ID parameter was provided.
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		renderError(w, http.StatusBadRequest, "Bad Request: Missing artist ID")
		return
	}

	// Validate that the ID is a valid integer.
	id, err := strconv.Atoi(idStr)
	if err != nil {
		renderError(w, http.StatusBadRequest, "Bad Request: Invalid artist ID")
		return
	}

	// Quick lower-bound check: IDs start at 1.
	// We intentionally skip an upper-bound check so the API can grow beyond 52 artists.
	if id < 1 {
		renderError(w, http.StatusNotFound, "Artist Not Found")
		return
	}

	// Fetch all artists to find the one matching the requested ID.
	artists, err := api.GetArtists()
	if err != nil {
		log.Printf("ArtistHandler: failed to fetch artists: %v", err)
		renderError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// Search for the artist by ID.
	var found *models.Artist
	for i := range artists {
		if artists[i].ID == id {
			found = &artists[i]
			break
		}
	}

	if found == nil {
		renderError(w, http.StatusNotFound, "Artist Not Found")
		return
	}

	// Fetch the relation data (locations → dates mapping) for this artist.
	relation, err := api.FetchRelation(id)
	if err != nil {
		log.Printf("ArtistHandler: failed to fetch relation for ID %d: %v", id, err)
		renderError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	data := models.ArtistPage{
		Artist:   *found,
		Relation: relation,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, "artist.html", data); err != nil {
		log.Printf("ArtistHandler: template execution error: %v", err)
	}
}

// SearchHandler handles the client-server event: live search.
// It receives a query parameter "q", filters artists by name (case-insensitive),
// and returns matching artists as a JSON array.
// This is the required "client call to the server" (request-response event).
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method Not Allowed"})
		return
	}

	query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))

	artists, err := api.GetArtists()
	if err != nil {
		log.Printf("SearchHandler: failed to fetch artists: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Internal Server Error"})
		return
	}

	// Filter artists whose name contains the search query.
	var results []models.Artist
	if query == "" {
		results = artists
	} else {
		for _, a := range artists {
			if strings.Contains(strings.ToLower(a.Name), query) {
				results = append(results, a)
			}
		}
	}

	// Ensure we always return an array, never null.
	if results == nil {
		results = []models.Artist{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		log.Printf("SearchHandler: JSON encode error: %v", err)
	}
}

// renderError renders the error.html template with the given status code and message.
// This ensures the site never shows a raw error — every error gets a styled page.
func renderError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)

	data := models.ErrorData{
		StatusCode: statusCode,
		Message:    message,
	}

	if err := templates.ExecuteTemplate(w, "error.html", data); err != nil {
		// Last resort: if even the error template fails, write plain text.
		log.Printf("renderError: template execution failed: %v", err)
		http.Error(w, message, statusCode)
	}
}

// FormatLocation converts an API location string like "north_carolina-usa"
// into a human-readable format like "North Carolina, USA".
func FormatLocation(location string) string {
	parts := strings.Split(location, "-")
	for i, part := range parts {
		part = strings.ReplaceAll(part, "_", " ")
		upper := strings.ToUpper(part)
		// Keep common country abbreviations uppercase.
		if upper == "USA" || upper == "UK" {
			parts[i] = upper
		} else {
			parts[i] = titleCase(part)
		}
	}
	return strings.Join(parts, ", ")
}

// titleCase converts a string to Title Case without using the deprecated strings.Title.
func titleCase(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		runes := []rune(word)
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
			for j := 1; j < len(runes); j++ {
				runes[j] = unicode.ToLower(runes[j])
			}
		}
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}
