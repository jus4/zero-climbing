package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/jus4/zero-climbing/db"
	"github.com/jus4/zero-climbing/db/sqlc"
	"github.com/jus4/zero-climbing/handlers"
	"github.com/jus4/zero-climbing/internal/model"
	"github.com/jus4/zero-climbing/services"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const defaultPort = "8082"

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("Error loading .env file")
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		log.Fatal("SESSION_SECRET is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	handlers.Configure(sessionSecret, handlers.CookieConfig{
		Secure:   strings.EqualFold(os.Getenv("COOKIE_SECURE"), "true"),
		Domain:   os.Getenv("COOKIE_DOMAIN"),
		SameSite: parseSameSite(os.Getenv("COOKIE_SAMESITE")),
	})

	e := echo.New()

	e.Static("/static", "assets")

	e.HTTPErrorHandler = handlers.CustomHTTPErrorHandler

	e.Use(middleware.Logger())

	e.Use(session.Middleware(sessions.NewCookieStore([]byte(sessionSecret))))

	pool, err := db.NewPgxPool()
	if err != nil {
		e.Logger.Fatalf("failed to connect to database: %s", err)
	}
	defer pool.Close()

	queries := sqlc.New(pool)

	oauthCfg := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	um := model.NewUserModel(queries)
	ah := handlers.NewAuthHandler(um, oauthCfg)

	ls := services.NewLocationServices(pool, queries)
	ih := handlers.NewImportHandler(ls)
	lh := handlers.NewLocationHandler(ls)

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	apiOAuthCfg := *oauthCfg
	apiOAuthCfg.RedirectURL = os.Getenv("GOOGLE_API_REDIRECT_URL")
	if apiOAuthCfg.RedirectURL == "" {
		apiOAuthCfg.RedirectURL = "http://localhost:8082/api/v1/auth/google/callback"
	}

	api := handlers.NewAPIHandler(um, ls, ls, &apiOAuthCfg, frontendURL)

	handlers.SetupRoutes(e, ah, ih, lh, api)

	go func() {
		if err := e.Start(":" + port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatalf("shutting down the server: %s", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

func parseSameSite(value string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
