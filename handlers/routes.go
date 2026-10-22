package handlers

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/patrickmn/go-cache"
)

var tileCache = cache.New(30*time.Minute, 1*time.Hour)

func SetupRoutes(e *echo.Echo, api *APIHandler) {
	e.GET("/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	csrf := middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "header:X-CSRF-Token",
		CookieName:     "csrf_token",
		CookiePath:     "/",
		CookieDomain:   cookieConfig.Domain,
		CookieSecure:   cookieConfig.Secure,
		CookieSameSite: cookieConfig.SameSite,
		ErrorHandler: func(err error, c echo.Context) error {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "invalid csrf token"})
		},
	})

	apiV1 := e.Group("/api/v1", csrf)
	apiV1.GET("/auth/me", api.Me)
	apiV1.GET("/auth/google", api.GoogleLogin)
	apiV1.GET("/auth/google/callback", api.GoogleCallback)
	apiV1.POST("/auth/username", api.Username)
	apiV1.POST("/auth/logout", api.Logout)
	apiV1.GET("/locations", api.Locations)
	apiV1.PATCH("/locations/:id", api.UpdateLocation)
	apiV1.DELETE("/locations/:id", api.DeleteLocation)
	apiV1.POST("/import", api.ImportUpload, middleware.BodyLimit("50M"))
	apiV1.POST("/import/confirm", api.ImportConfirm)

	e.GET("/proxy/wmts/*", func(c echo.Context) error {
		tilePath := c.Param("*")
		cacheKey := tilePath
		mapApiKey := os.Getenv("MAP_API_KEY")

		if cached, found := tileCache.Get(cacheKey); found {
			cachedData := cached.(map[string]any)
			body := cachedData["body"].([]byte)
			contentType := cachedData["contentType"].(string)

			return c.Blob(http.StatusOK, contentType, body)
		}

		targetURL := "https://avoin-karttakuva.maanmittauslaitos.fi/avoin/wmts/" + tilePath + "?api-key=" + mapApiKey

		resp, err := http.Get(targetURL)
		if err != nil {
			return c.String(http.StatusBadGateway, "Failed to fetch tile")
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to read tile response")
		}

		tileCache.Set(cacheKey, map[string]any{
			"body":        bodyBytes,
			"contentType": resp.Header.Get("Content-Type"),
		}, cache.DefaultExpiration)

		return c.Blob(resp.StatusCode, resp.Header.Get("Content-Type"), bodyBytes)
	})
}
