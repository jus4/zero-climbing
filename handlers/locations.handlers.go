package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jus4/zero-climbing/services"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type LocationService interface {
	GetPointLocations(ctx context.Context) ([]services.Location, error)
}

func NewLocationHandler(locations LocationService) *LocationHandler {
	return &LocationHandler{
		locations: locations,
	}
}

type LocationHandler struct {
	locations LocationService
}

func (lh *LocationHandler) mapLocationsHandler(c echo.Context) error {
	if !hasSessionCookie(c) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	locations, err := lh.locations.GetPointLocations(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not load locations"})
	}

	features := make([]map[string]any, 0, len(locations))
	for _, loc := range locations {
		features = append(features, map[string]any{
			"type":       "Feature",
			"id":         loc.ID,
			"properties": map[string]any{"title": loc.Title, "description": loc.Description, "user": loc.Username},
			"geometry":   json.RawMessage(loc.Geojson),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"type":     "FeatureCollection",
		"features": features,
	})
}

func hasSessionCookie(c echo.Context) bool {
	sess, err := session.Get(auth_sessions_key, c)
	if err != nil {
		return false
	}
	return !sess.IsNew
}
