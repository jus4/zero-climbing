package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/jus4/zero-climbing/views/errors_pages"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

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

	if strings.HasPrefix(c.Request().URL.Path, "/api/") {
		c.JSON(code, map[string]string{"error": message})
		return
	}

	/* errorPage := fmt.Sprintf("views/%d.html", code)
	if err := c.File(errorPage); err != nil {
		c.Logger().Error(err)
	} */

	var errorPage func(fp bool) templ.Component

	switch code {
	case 401:
		errorPage = errors_pages.Error401
	case 404:
		errorPage = errors_pages.Error404
	case 405:
		errorPage = errors_pages.Error404
	case 500:
		errorPage = errors_pages.Error500
	default:
		errorPage = errors_pages.Error500
	}

	// isError = true
	c.Set("ISERROR", true)

	fromProtected, _ := c.Get("FROMPROTECTED").(bool)

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	c.Response().WriteHeader(code)

	renderView(c, errors_pages.ErrorIndex(
		fmt.Sprintf("| Error (%d)", code),
		"",
		"",
		fromProtected,
		c.Get("ISERROR").(bool),
		errorPage(fromProtected),
	))
}

func RouteNotFoundHandler(c echo.Context) error {
	if strings.HasPrefix(c.Request().URL.Path, "/api/") {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	c.Response().WriteHeader(http.StatusNotFound)

	return renderView(c, errors_pages.ErrorIndex(
		fmt.Sprintf("| Error (%d)", 404),
		"",
		"",
		false,
		true,
		errors_pages.Error404(false),
	))
}
