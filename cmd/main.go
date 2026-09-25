package main

import (
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/jus4/zero-climbing/db"
	"github.com/jus4/zero-climbing/db/sqlc"
	"github.com/jus4/zero-climbing/handlers"
	"github.com/jus4/zero-climbing/services"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
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

	e.Use(session.Middleware(sessions.NewCookieStore([]byte(SECRET_KEY))))

	pool, err := db.NewPgxPool()
	if err != nil {
		e.Logger.Fatalf("failed to connect to database: %s", err)
	}
	defer pool.Close()

	queries := sqlc.New(pool)

	us := services.NewUserServices(queries)
	ah := handlers.NewAuthHandler(us)

	handlers.SetupRoutes(e, ah)

	e.Logger.Fatal(e.Start(":8082"))
}
