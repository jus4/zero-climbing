package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/jus4/zero-climbing/internal/model"
	"github.com/jus4/zero-climbing/views/auth_views"

	"github.com/a-h/templ"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

const (
	auth_sessions_key string = "authenticate-sessions"
	auth_key          string = "authenticated"
	user_id_key       string = "user_id"
	alias_key         string = "alias"
	email_key         string = "email"
	tzone_key         string = "time_zone"
	oauth_state_key   string = "oauth_state"
	pending_email_key string = "pending_email"
)

type AuthService interface {
	GetByEmail(ctx context.Context, email string) (model.User, error)
	GetByAlias(ctx context.Context, alias string) (model.User, error)
	Create(ctx context.Context, email, alias string) (model.User, error)
}

func NewAuthHandler(us AuthService, oauthCfg *oauth2.Config) *AuthHandler {
	return &AuthHandler{
		UserModel: us,
		oauthCfg:  oauthCfg,
	}
}

type AuthHandler struct {
	UserModel AuthService
	oauthCfg  *oauth2.Config
}

func (ah *AuthHandler) homeHandler(c echo.Context) error {
	fromProtected, ok := c.Get("FROMPROTECTED").(bool)
	if !ok {
		return errors.New("invalid type for key 'FROMPROTECTED'")
	}
	homeView := auth_views.Home(fromProtected)
	c.Set("ISERROR", false)

	username, email := userDisplay(c)

	return renderView(c, auth_views.HomeIndex(
		"| Home",
		username,
		email,
		fromProtected,
		c.Get("ISERROR").(bool),
		getFlashmessages(c, "error"),
		getFlashmessages(c, "success"),
		homeView,
	))
}

func (ah *AuthHandler) loginHandler(c echo.Context) error {
	fromProtected, ok := c.Get("FROMPROTECTED").(bool)
	if !ok {
		return errors.New("invalid type for key 'FROMPROTECTED'")
	}
	loginView := auth_views.Login(fromProtected)
	c.Set("ISERROR", false)

	username, email := userDisplay(c)

	return renderView(c, auth_views.LoginIndex(
		"| Login",
		username,
		email,
		fromProtected,
		c.Get("ISERROR").(bool),
		getFlashmessages(c, "error"),
		getFlashmessages(c, "success"),
		loginView,
	))
}

func (ah *AuthHandler) flagsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get(auth_sessions_key, c)

		if sess.IsNew {
			sess.Options = sessionCookieOptions(0)
			if err := sess.Save(c.Request(), c.Response()); err != nil {
				c.Logger().Error(err)
			}
		}

		if auth, ok := sess.Values[auth_key].(bool); !ok || !auth {
			c.Set("FROMPROTECTED", false)

			return next(c)
		}

		setUserContext(c, sess)

		c.Set("FROMPROTECTED", true)

		return next(c)
	}
}

func (ah *AuthHandler) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get(auth_sessions_key, c)
		if auth, ok := sess.Values[auth_key].(bool); !ok || !auth {
			c.Set("FROMPROTECTED", false)

			return echo.NewHTTPError(echo.ErrUnauthorized.Code, "Please provide valid credentials")
		}

		setUserContext(c, sess)

		c.Set("FROMPROTECTED", true)

		return next(c)
	}
}

func (ah *AuthHandler) logoutHandler(c echo.Context) error {
	sess, _ := session.Get(auth_sessions_key, c)
	sess.Values = map[interface{}]interface{}{
		auth_key:    false,
		user_id_key: "",
		alias_key:   "",
		email_key:   "",
		tzone_key:   "",
	}
	sess.Save(c.Request(), c.Response())

	setFlashmessages(c, "success", "You have successfully logged out!!")

	c.Set("FROMPROTECTED", false)

	return c.Redirect(http.StatusSeeOther, "/login")
}

func startSession(c echo.Context, user model.User, tzone string) error {
	sess, _ := session.Get(auth_sessions_key, c)
	sess.Options = sessionCookieOptions(3600)

	sess.Values = map[interface{}]interface{}{
		auth_key:    true,
		user_id_key: user.ID,
		alias_key:   user.Alias,
		email_key:   user.Email,
		tzone_key:   tzone,
	}

	return sess.Save(c.Request(), c.Response())
}

func setUserContext(c echo.Context, sess *sessions.Session) {
	if userId, ok := sess.Values[user_id_key].(int64); ok && userId != 0 {
		c.Set(user_id_key, userId)
	}

	if alias, ok := sess.Values[alias_key].(string); ok && len(alias) != 0 {
		c.Set(alias_key, alias)
	}

	if email, ok := sess.Values[email_key].(string); ok && len(email) != 0 {
		c.Set(email_key, email)
	}

	if tzone, ok := sess.Values[tzone_key].(string); ok && len(tzone) != 0 {
		c.Set(tzone_key, tzone)
	}
}

func userDisplay(c echo.Context) (string, string) {
	username, _ := c.Get(alias_key).(string)
	email, _ := c.Get(email_key).(string)

	return username, email
}

func renderView(c echo.Context, cmp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)

	return cmp.Render(c.Request().Context(), c.Response().Writer)
}
