package handlers

import (
	"encoding/json"
	"testing"

	"github.com/jus4/zero-climbing/services"
)

func TestBuildFeatureCollection(t *testing.T) {
	locations := []services.Location{
		{
			ID:          7,
			UserID:      3,
			Title:       "Peak",
			Description: "desc",
			IsPublic:    true,
			Username:    "alice",
			Geojson:     json.RawMessage(`{"type":"Point","coordinates":[1,2]}`),
		},
	}

	fc := buildFeatureCollection(locations)
	if fc["type"] != "FeatureCollection" {
		t.Fatalf("type = %v, want FeatureCollection", fc["type"])
	}

	features, ok := fc["features"].([]map[string]any)
	if !ok || len(features) != 1 {
		t.Fatalf("unexpected features: %#v", fc["features"])
	}

	if features[0]["id"] != int64(7) {
		t.Fatalf("id = %v, want 7", features[0]["id"])
	}

	props, ok := features[0]["properties"].(map[string]any)
	if !ok || props["is_public"] != true {
		t.Fatalf("is_public = %v, want true", props["is_public"])
	}
}

func TestBuildFeatureCollectionEmpty(t *testing.T) {
	fc := buildFeatureCollection(nil)
	features, ok := fc["features"].([]map[string]any)
	if !ok || len(features) != 0 {
		t.Fatalf("expected empty features, got %#v", fc["features"])
	}
}
