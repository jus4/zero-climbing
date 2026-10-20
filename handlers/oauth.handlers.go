package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jus4/zero-climbing/views/auth_views"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

const (
	googleUserInfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"
	minAliasLen       = 4
	maxAliasLen       = 64
)

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func (ah *AuthHandler) googleLoginHandler(c echo.Context) error {
	state, err := randomState()
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(echo.ErrInternalServerError.Code, "could not start sign-in")
	}

	sess, _ := session.Get(auth_sessions_key, c)
	sess.Values[oauth_state_key] = state
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(echo.ErrInternalServerError.Code, "could not start sign-in")
	}

	return c.Redirect(http.StatusSeeOther, ah.oauthCfg.AuthCodeURL(state))
}

func (ah *AuthHandler) googleCallbackHandler(c echo.Context) error {
	sess, _ := session.Get(auth_sessions_key, c)

	if c.QueryParam("error") != "" {
		setFlashmessages(c, "error", "Google sign-in was cancelled")
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	state := c.QueryParam("state")
	expected, _ := sess.Values[oauth_state_key].(string)
	delete(sess.Values, oauth_state_key)
	sess.Save(c.Request(), c.Response())

	if state == "" || expected == "" || state != expected {
		setFlashmessages(c, "error", "Invalid sign-in state, please try again")
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	code := c.QueryParam("code")
	if code == "" {
		setFlashmessages(c, "error", "Missing authorization code")
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	ctx := c.Request().Context()

	token, err := ah.oauthCfg.Exchange(ctx, code)
	if err != nil {
		c.Logger().Error(err)
		setFlashmessages(c, "error", "Could not complete Google sign-in")
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	info, err := fetchGoogleUserInfo(ctx, ah.oauthCfg, token)
	if err != nil {
		c.Logger().Error(err)
		setFlashmessages(c, "error", "Could not load your Google profile")
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	if info.Email == "" || !info.EmailVerified {
		setFlashmessages(c, "error", "Your Google account has no verified email")
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	user, err := ah.UserModel.GetByEmail(ctx, info.Email)
	if err == nil {
		if err := startSession(c, user, c.Request().Header.Get("X-Timezone")); err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(echo.ErrInternalServerError.Code, "could not start session")
		}

		setFlashmessages(c, "success", "You have successfully logged in!!")

		return c.Redirect(http.StatusSeeOther, "/")
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		c.Logger().Error(err)
		setFlashmessages(c, "error", "Could not complete Google sign-in")
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	sess, _ = session.Get(auth_sessions_key, c)
	sess.Values[pending_email_key] = info.Email
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(echo.ErrInternalServerError.Code, "could not start session")
	}

	return c.Redirect(http.StatusSeeOther, "/onboarding")
}

func (ah *AuthHandler) onboardingPageHandler(c echo.Context) error {
	fromProtected, ok := c.Get("FROMPROTECTED").(bool)
	if !ok {
		return errors.New("invalid type for key 'FROMPROTECTED'")
	}

	sess, _ := session.Get(auth_sessions_key, c)
	email, _ := sess.Values[pending_email_key].(string)
	if email == "" {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	username, _ := userDisplay(c)
	c.Set("ISERROR", false)

	return renderView(c, auth_views.OnboardingIndex(
		"| Choose a username",
		username,
		email,
		fromProtected,
		c.Get("ISERROR").(bool),
		getFlashmessages(c, "error"),
		getFlashmessages(c, "success"),
		auth_views.Onboarding(email, fromProtected),
	))
}

func (ah *AuthHandler) onboardingSubmitHandler(c echo.Context) error {
	sess, _ := session.Get(auth_sessions_key, c)
	email, _ := sess.Values[pending_email_key].(string)
	if email == "" {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	username := strings.TrimSpace(c.FormValue("username"))
	if len(username) < minAliasLen || len(username) > maxAliasLen {
		setFlashmessages(c, "error", fmt.Sprintf(
			"Username must be between %d and %d characters",
			minAliasLen,
			maxAliasLen,
		))

		return c.Redirect(http.StatusSeeOther, "/onboarding")
	}

	ctx := c.Request().Context()

	if _, err := ah.UserModel.GetByAlias(ctx, username); err == nil {
		setFlashmessages(c, "error", "That username is already taken")

		return c.Redirect(http.StatusSeeOther, "/onboarding")
	} else if !errors.Is(err, pgx.ErrNoRows) {
		c.Logger().Error(err)
		return echo.NewHTTPError(echo.ErrInternalServerError.Code, "could not check username")
	}

	user, err := ah.UserModel.Create(ctx, email, username)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			setFlashmessages(c, "error", "That username is already taken")

			return c.Redirect(http.StatusSeeOther, "/onboarding")
		}

		c.Logger().Error(err)
		return echo.NewHTTPError(echo.ErrInternalServerError.Code, "could not create user")
	}

	if err := startSession(c, user, c.Request().Header.Get("X-Timezone")); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(echo.ErrInternalServerError.Code, "could not start session")
	}

	setFlashmessages(c, "success", "Welcome! Your account is ready.")

	return c.Redirect(http.StatusSeeOther, "/")
}

func fetchGoogleUserInfo(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (googleUserInfo, error) {
	client := cfg.Client(ctx, token)

	resp, err := client.Get(googleUserInfoURL)
	if err != nil {
		return googleUserInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return googleUserInfo{}, fmt.Errorf("google userinfo: unexpected status %d", resp.StatusCode)
	}

	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return googleUserInfo{}, err
	}

	return info, nil
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
