package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
)

// RegisterStatic serves the built SPA from distDir, falling back to index.html
// for client-side routes. If distDir has no index.html (e.g. local API-only
// development), it logs a warning and registers nothing.
func RegisterStatic(e *echo.Echo, distDir string) {
	if distDir == "" {
		distDir = "/app/dist"
	}

	if _, err := os.Stat(filepath.Join(distDir, "index.html")); err != nil {
		e.Logger.Warnf("static dir %q not found; serving API only", distDir)
		return
	}

	fileServer := http.FileServer(http.Dir(distDir))

	serve := func(c echo.Context) error {
		path := c.Request().URL.Path

		// Unmatched API/proxy paths should stay JSON 404s, not index.html.
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/proxy/") {
			return echo.NewHTTPError(http.StatusNotFound, "Not Found")
		}

		filePath := filepath.Join(distDir, filepath.Clean("/"+path))
		if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(c.Response(), c.Request())
			return nil
		}

		return c.File(filepath.Join(distDir, "index.html"))
	}

	e.GET("/", serve)
	e.GET("/*", serve)
}
