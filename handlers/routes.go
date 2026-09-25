package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"io"
	"net/http"
	"os"
	"time"
)

func SetupRoutes(e *echo.Echo, ah *AuthHandler) {
	e.GET("/", ah.flagsMiddleware(ah.homeHandler))

	e.POST("/logout", ah.authMiddleware(ah.logoutHandler))

	e.GET("/proxy/wmts/*", func(c echo.Context) error {
		var tileCache = cache.New(30*time.Minute, 1*time.Hour)
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

	/* ↓ Fallback Page ↓ */
	e.GET("/*", RouteNotFoundHandler)
}
