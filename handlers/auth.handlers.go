package handlers

import (
	"context"

	"github.com/jus4/zero-climbing/internal/model"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
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
