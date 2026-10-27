package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jus4/zero-climbing/services"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"golang.org/x/oauth2"
)

// APIHandler serves the versioned JSON API consumed by the SPA.
type APIHandler struct {
	users       AuthService
	locations   LocationService
	imports     ImportService
	oauthCfg    *oauth2.Config
	frontendURL string
}

func NewAPIHandler(users AuthService, locations LocationService, imports ImportService, oauthCfg *oauth2.Config, frontendURL string) *APIHandler {
	return &APIHandler{
		users:       users,
		locations:   locations,
		imports:     imports,
		oauthCfg:    oauthCfg,
		frontendURL: frontendURL,
	}
}

func (h *APIHandler) Me(c echo.Context) error {
	sess, err := session.Get(auth_sessions_key, c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	if auth, ok := sess.Values[auth_key].(bool); ok && auth {
		userID, _ := sess.Values[user_id_key].(int64)
		alias, _ := sess.Values[alias_key].(string)
		email, _ := sess.Values[email_key].(string)

		return c.JSON(http.StatusOK, map[string]any{
			"id":                 userID,
			"alias":              alias,
			"email":              email,
			"onboardingRequired": false,
		})
	}

	if email, ok := sess.Values[pending_email_key].(string); ok && email != "" {
		return c.JSON(http.StatusOK, map[string]any{
			"id":                 0,
			"alias":              "",
			"email":              email,
			"onboardingRequired": true,
		})
	}

	return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
}

func (h *APIHandler) GoogleLogin(c echo.Context) error {
	state, err := randomState()
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "could not start sign-in")
	}

	sess, _ := session.Get(auth_sessions_key, c)
	sess.Values[oauth_state_key] = state
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "could not start sign-in")
	}

	return c.Redirect(http.StatusFound, h.oauthCfg.AuthCodeURL(state))
}

func (h *APIHandler) GoogleCallback(c echo.Context) error {
	sess, _ := session.Get(auth_sessions_key, c)

	if c.QueryParam("error") != "" {
		return c.Redirect(http.StatusFound, h.frontendErrorURL("Google sign-in was cancelled"))
	}

	state := c.QueryParam("state")
	expected, _ := sess.Values[oauth_state_key].(string)
	delete(sess.Values, oauth_state_key)
	sess.Save(c.Request(), c.Response())

	if state == "" || expected == "" || state != expected {
		return c.Redirect(http.StatusFound, h.frontendErrorURL("Invalid sign-in state, please try again"))
	}

	code := c.QueryParam("code")
	if code == "" {
		return c.Redirect(http.StatusFound, h.frontendErrorURL("Missing authorization code"))
	}

	ctx := c.Request().Context()

	token, err := h.oauthCfg.Exchange(ctx, code)
	if err != nil {
		c.Logger().Error(err)
		return c.Redirect(http.StatusFound, h.frontendErrorURL("Could not complete Google sign-in"))
	}

	info, err := fetchGoogleUserInfo(ctx, h.oauthCfg, token)
	if err != nil {
		c.Logger().Error(err)
		return c.Redirect(http.StatusFound, h.frontendErrorURL("Could not load your Google profile"))
	}

	if info.Email == "" || !info.EmailVerified {
		return c.Redirect(http.StatusFound, h.frontendErrorURL("Your Google account has no verified email"))
	}

	user, err := h.users.GetByEmail(ctx, info.Email)
	if err == nil {
		if err := startSession(c, user, c.Request().Header.Get("X-Timezone")); err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError, "could not start session")
		}

		return c.Redirect(http.StatusFound, h.frontendURL)
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		c.Logger().Error(err)
		return c.Redirect(http.StatusFound, h.frontendErrorURL("Could not complete Google sign-in"))
	}

	sess, _ = session.Get(auth_sessions_key, c)
	sess.Values[pending_email_key] = info.Email
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "could not start session")
	}

	return c.Redirect(http.StatusFound, h.frontendURL)
}

func (h *APIHandler) Username(c echo.Context) error {
	sess, _ := session.Get(auth_sessions_key, c)
	email, _ := sess.Values[pending_email_key].(string)
	if email == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "no pending sign-in"})
	}

	var req struct {
		Username string `json:"username"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	username := strings.TrimSpace(req.Username)
	if len(username) < minAliasLen || len(username) > maxAliasLen {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("username must be between %d and %d characters", minAliasLen, maxAliasLen),
		})
	}

	ctx := c.Request().Context()

	if _, err := h.users.GetByAlias(ctx, username); err == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "that username is already taken"})
	} else if !errors.Is(err, pgx.ErrNoRows) {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "could not check username")
	}

	user, err := h.users.Create(ctx, email, username)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return c.JSON(http.StatusConflict, map[string]string{"error": "that username is already taken"})
		}

		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "could not create user")
	}

	if err := startSession(c, user, c.Request().Header.Get("X-Timezone")); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "could not start session")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"id":                 user.ID,
		"alias":              user.Alias,
		"email":              user.Email,
		"onboardingRequired": false,
	})
}

func (h *APIHandler) Logout(c echo.Context) error {
	sess, _ := session.Get(auth_sessions_key, c)
	sess.Values = map[interface{}]interface{}{
		auth_key:    false,
		user_id_key: "",
		alias_key:   "",
		email_key:   "",
		tzone_key:   "",
	}
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "could not log out")
	}

	return c.JSON(http.StatusOK, map[string]bool{"ok": true})
}

func (h *APIHandler) Locations(c echo.Context) error {
	locations, err := h.locations.GetPublicPointLocations(c.Request().Context())
	if err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not load locations"})
	}

	return c.JSON(http.StatusOK, buildFeatureCollection(locations))
}

func (h *APIHandler) PrivateLocations(c echo.Context) error {
	userID, _, ok := authenticatedUser(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	locations, err := h.locations.GetPrivatePointLocations(c.Request().Context(), userID)
	if err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not load locations"})
	}

	return c.JSON(http.StatusOK, buildFeatureCollection(locations))
}

func (h *APIHandler) UpdateLocation(c echo.Context) error {
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
		IsPublic    bool   `json:"is_public"`
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

	if err := h.locations.UpdateLocation(c.Request().Context(), id, userID, title, description, req.IsPublic); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "location not found"})
		}

		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not update location"})
	}

	return c.JSON(http.StatusOK, map[string]bool{"ok": true})
}

func (h *APIHandler) DeleteLocation(c echo.Context) error {
	userID, _, ok := authenticatedUser(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid location id"})
	}

	if err := h.locations.DeleteLocation(c.Request().Context(), id, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "location not found"})
		}

		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not delete location"})
	}

	return c.JSON(http.StatusOK, map[string]bool{"ok": true})
}

func (h *APIHandler) ImportUpload(c echo.Context) error {
	if _, _, ok := authenticatedUser(c); !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "please choose a file to upload"})
	}

	if file.Size > maxUploadSize {
		return c.JSON(http.StatusRequestEntityTooLarge, map[string]string{"error": "file is too large (max 50 MB)"})
	}

	tmp, err := os.CreateTemp("", "zero-climbing-import-*")
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "could not process upload")
	}
	defer os.Remove(tmp.Name())

	src, err := file.Open()
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "could not read the uploaded file")
	}
	defer src.Close()

	if _, err := io.Copy(tmp, src); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "could not process upload")
	}

	candidates, err := services.ParseImportFile(tmp.Name(), file.Filename)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	token := newImportToken()
	importStore.Set(token, candidates, cache.DefaultExpiration)

	return c.JSON(http.StatusOK, map[string]any{
		"token":      token,
		"candidates": candidates,
	})
}

func (h *APIHandler) ImportConfirm(c echo.Context) error {
	userID, _, ok := authenticatedUser(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req struct {
		Token     string            `json:"token"`
		Selection []importSelection `json:"selection"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	candidates, ok := getCandidates(req.Token)
	if !ok {
		return c.JSON(http.StatusGone, map[string]string{"error": "import session expired or not found, please upload the file again"})
	}

	toImport := selectCandidates(candidates, req.Selection)
	if len(toImport) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "no items selected"})
	}

	imported, skipped, err := h.imports.ImportLocations(c.Request().Context(), userID, toImport)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "import failed")
	}

	importStore.Delete(req.Token)

	return c.JSON(http.StatusOK, map[string]int{"imported": imported, "skipped": skipped})
}

func (h *APIHandler) frontendErrorURL(message string) string {
	u, err := url.Parse(h.frontendURL)
	if err != nil {
		return h.frontendURL
	}

	u.Path = strings.TrimSuffix(u.Path, "/") + "/login"
	q := u.Query()
	q.Set("error", message)
	u.RawQuery = q.Encode()

	return u.String()
}
