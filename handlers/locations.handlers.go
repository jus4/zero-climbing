package handlers

import (
	"context"
	"encoding/json"

	"github.com/jus4/zero-climbing/services"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

const (
	maxTitleLen       = 200
	maxDescriptionLen = 2000
)

type LocationService interface {
	GetPublicPointLocations(ctx context.Context) ([]services.Location, error)
	GetPrivatePointLocations(ctx context.Context, userID int64) ([]services.Location, error)
	UpdateLocation(ctx context.Context, id, userID int64, title, description string, isPublic bool) error
	DeleteLocation(ctx context.Context, id, userID int64) error
}

func buildFeatureCollection(locations []services.Location) map[string]any {
	features := make([]map[string]any, 0, len(locations))
	for _, loc := range locations {
		features = append(features, map[string]any{
			"type": "Feature",
			"id":   loc.ID,
			"properties": map[string]any{
				"title":       loc.Title,
				"description": loc.Description,
				"is_public":   loc.IsPublic,
				"user":        loc.Username,
				"user_id":     loc.UserID,
			},
			"geometry": json.RawMessage(loc.Geojson),
		})
	}

	return map[string]any{
		"type":     "FeatureCollection",
		"features": features,
	}
}

func authenticatedUser(c echo.Context) (int64, string, bool) {
	sess, err := session.Get(auth_sessions_key, c)
	if err != nil {
		return 0, "", false
	}

	if auth, ok := sess.Values[auth_key].(bool); !ok || !auth {
		return 0, "", false
	}

	userID, ok := sess.Values[user_id_key].(int64)
	if !ok || userID == 0 {
		return 0, "", false
	}

	alias, _ := sess.Values[alias_key].(string)

	return userID, alias, true
}
