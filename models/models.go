// Package models defines the data structures used throughout the application.
// These structs map directly to the JSON responses from the Groupie Trackers API.
package models

// Artist represents a band or solo artist from the API.
type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Locations    string   `json:"locations"`
	ConcertDates string   `json:"concertDates"`
	Relations    string   `json:"relations"`
}

// RelationData links concert locations to their dates for a single artist.
type RelationData struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

// ArtistPage holds combined data for rendering the artist detail page.
type ArtistPage struct {
	Artist   Artist
	Relation RelationData
}

// ErrorData holds information for rendering the error page.
type ErrorData struct {
	StatusCode int
	Message    string
}
