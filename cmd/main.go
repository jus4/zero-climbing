package main

import (
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
	"log"
	"os"
)

const (
	SECRET_KEY string = "secret"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Print("Error loading .env file")
	}
	e := echo.New()

	e.Static("/static", "assets")

	e.HTTPErrorHandler = handlers.CustomHTTPErrorHandler

	e.Use(middleware.Logger())

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		sessionSecret = SECRET_KEY
	}
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

	e.Logger.Fatal(e.Start(":8082"))
}
