package api

import (
	"testing"
)

// TestFetchArtists verifies that the API returns all 52 artists
// and that the first artist is Queen (ID 1).
func TestFetchArtists(t *testing.T) {
	artists, err := GetArtists()
	if err != nil {
		t.Fatalf("GetArtists() returned error: %v", err)
	}

	if len(artists) == 0 {
		t.Fatal("GetArtists() returned empty slice")
	}

	// The API should return 52 artists.
	if len(artists) != 52 {
		t.Errorf("GetArtists() returned %d artists, want 52", len(artists))
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

// TestFetchRelation verifies that fetching a relation by valid ID
// returns a non-empty DatesLocations map.
func TestFetchRelation(t *testing.T) {
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
}

// TestFetchRelationInvalidID verifies that a non-existent ID
// returns either an error or empty relation data.
func TestFetchRelationInvalidID(t *testing.T) {
	relation, err := FetchRelation(99999)
	// The external API may return 200 with empty data for non-existent IDs.
	// We verify that the result is at least empty/unusable.
	if err != nil {
		// Error is the expected good behavior — test passes.
		return
	}
	if len(relation.DatesLocations) > 0 {
		t.Error("FetchRelation(99999) should return empty DatesLocations for invalid ID")
	}
}
