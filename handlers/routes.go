package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/patrickmn/go-cache"
	"io"
	"net/http"
	"os"
	"time"
)

func SetupRoutes(e *echo.Echo, ah *AuthHandler, ih *ImportHandler, lh *LocationHandler, api *APIHandler) {
	e.GET("/", ah.flagsMiddleware(ah.homeHandler))

	csrf := middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "header:X-CSRF-Token",
		CookieName:     "csrf_token",
		CookiePath:     "/",
		CookieSameSite: http.SameSiteLaxMode,
		ErrorHandler: func(err error, c echo.Context) error {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "invalid csrf token"})
		},
	})

	e.GET("/api/locations", csrf(lh.mapLocationsHandler))
	e.GET("/api/me", lh.meHandler)
	e.POST("/api/locations/:id/edit", csrf(lh.updateLocationHandler))
	e.DELETE("/api/locations/:id", csrf(lh.deleteLocationHandler))

	apiV1 := e.Group("/api/v1", csrf)
	apiV1.GET("/auth/me", api.Me)
	apiV1.GET("/auth/google", api.GoogleLogin)
	apiV1.GET("/auth/google/callback", api.GoogleCallback)
	apiV1.POST("/auth/username", api.Username)
	apiV1.POST("/auth/logout", api.Logout)
	apiV1.GET("/locations", api.Locations)
	apiV1.PATCH("/locations/:id", api.UpdateLocation)
	apiV1.DELETE("/locations/:id", api.DeleteLocation)

	e.GET("/login", ah.flagsMiddleware(ah.loginHandler))

	e.GET("/auth/google", ah.flagsMiddleware(ah.googleLoginHandler))
	e.GET("/auth/google/callback", ah.googleCallbackHandler)

	e.GET("/onboarding", ah.flagsMiddleware(ah.onboardingPageHandler))
	e.POST("/onboarding", ah.flagsMiddleware(ah.onboardingSubmitHandler))

	e.POST("/logout", ah.authMiddleware(ah.logoutHandler))

	e.GET("/import", ah.authMiddleware(ih.importPageHandler))
	e.POST("/import", ah.authMiddleware(ih.importUploadHandler), middleware.BodyLimit("50M"))
	e.GET("/import/preview", ah.authMiddleware(ih.importPreviewHandler))
	e.POST("/import/confirm", ah.authMiddleware(ih.importConfirmHandler))

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
