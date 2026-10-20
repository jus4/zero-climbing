package handlers

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// CookieConfig controls the attributes of the session and CSRF cookies so the
// same binary can run locally (Lax, insecure) and in production (Secure, with a
// shared domain) without code changes.
type CookieConfig struct {
	Secure   bool
	Domain   string
	SameSite http.SameSite
}

var (
	sessionSecret []byte
	cookieConfig  = CookieConfig{SameSite: http.SameSiteLaxMode}
)

// Configure stores the process-wide session secret and cookie settings. It must
// be called before routes are registered.
func Configure(secret string, cookies CookieConfig) {
	sessionSecret = []byte(secret)
	cookieConfig = cookies
}

func sessionCookieOptions(maxAge int) *sessions.Options {
	return &sessions.Options{
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   cookieConfig.Secure,
		Domain:   cookieConfig.Domain,
		SameSite: cookieConfig.SameSite,
	}
}
