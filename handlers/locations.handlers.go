package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jus4/zero-climbing/services"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

const (
	maxTitleLen       = 200
	maxDescriptionLen = 2000
)

type LocationService interface {
	GetPointLocations(ctx context.Context) ([]services.Location, error)
	UpdateLocation(ctx context.Context, id, userID int64, title, description string) error
	DeleteLocation(ctx context.Context, id, userID int64) error
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

	return c.JSON(http.StatusOK, buildFeatureCollection(locations))
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

func (lh *LocationHandler) meHandler(c echo.Context) error {
	userID, alias, ok := authenticatedUser(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"id":    userID,
		"alias": alias,
	})
}

func (lh *LocationHandler) updateLocationHandler(c echo.Context) error {
	userID, _, ok := authenticatedUser(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid location id"})
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "title cannot be empty"})
	}
	if len(title) > maxTitleLen {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "title is too long"})
	}
	description := strings.TrimSpace(req.Description)
	if len(description) > maxDescriptionLen {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "description is too long"})
	}

	if err := lh.locations.UpdateLocation(c.Request().Context(), id, userID, title, description); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "location not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not update location"})
	}

	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
}

func (lh *LocationHandler) deleteLocationHandler(c echo.Context) error {
	userID, _, ok := authenticatedUser(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid location id"})
	}

	if err := lh.locations.DeleteLocation(c.Request().Context(), id, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "location not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not delete location"})
	}

	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
}

func hasSessionCookie(c echo.Context) bool {
	sess, err := session.Get(auth_sessions_key, c)
	if err != nil {
		return false
	}
	return !sess.IsNew
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
