package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// CustomHTTPErrorHandler returns JSON errors for the API, which is the only
// surface served by this binary (the SPA and its static assets are served by
// nginx).
func CustomHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	code := http.StatusInternalServerError
	message := http.StatusText(code)
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		switch m := he.Message.(type) {
		case string:
			message = m
		case nil:
			message = http.StatusText(code)
		default:
			message = fmt.Sprintf("%v", m)
		}
	}
	c.Logger().Error(err)

	c.JSON(code, map[string]string{"error": message})
}
