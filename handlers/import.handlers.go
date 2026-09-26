package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jus4/zero-climbing/services"
	"github.com/jus4/zero-climbing/views/auth_views"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
)

var importStore = cache.New(30*time.Minute, 1*time.Hour)

const maxUploadSize = 50 << 20

type ImportService interface {
	ImportLocations(ctx context.Context, userID int64, items []services.ImportCandidate) error
}

func NewImportHandler(locations ImportService) *ImportHandler {
	return &ImportHandler{
		locations: locations,
	}
}

type ImportHandler struct {
	locations ImportService
}

func (ih *ImportHandler) importPageHandler(c echo.Context) error {
	fromProtected, ok := c.Get("FROMPROTECTED").(bool)
	if !ok {
		return errors.New("invalid type for key 'FROMPROTECTED'")
	}
	importView := auth_views.ImportForm(fromProtected)
	c.Set("ISERROR", false)

	return renderView(c, auth_views.ImportFormIndex(
		"| Import",
		"",
		fromProtected,
		c.Get("ISERROR").(bool),
		getFlashmessages(c, "error"),
		getFlashmessages(c, "success"),
		importView,
	))
}

func (ih *ImportHandler) importUploadHandler(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		setFlashmessages(c, "error", "Please choose a file to upload")
		return c.Redirect(http.StatusSeeOther, "/import")
	}

	if file.Size > maxUploadSize {
		setFlashmessages(c, "error", "File is too large (max 50 MB)")
		return c.Redirect(http.StatusSeeOther, "/import")
	}

	tmp, err := os.CreateTemp("", "zero-climbing-import-*")
	if err != nil {
		return echo.NewHTTPError(echo.ErrInternalServerError.Code, "could not process upload")
	}
	defer os.Remove(tmp.Name())

	src, err := file.Open()
	if err != nil {
		setFlashmessages(c, "error", "Could not read the uploaded file")
		return c.Redirect(http.StatusSeeOther, "/import")
	}
	defer src.Close()

	if _, err := io.Copy(tmp, src); err != nil {
		return echo.NewHTTPError(echo.ErrInternalServerError.Code, "could not process upload")
	}

	candidates, err := services.ParseImportFile(tmp.Name(), file.Filename)
	if err != nil {
		setFlashmessages(c, "error", err.Error())
		return c.Redirect(http.StatusSeeOther, "/import")
	}

	token := newImportToken()
	importStore.Set(token, candidates, cache.DefaultExpiration)

	return c.Redirect(http.StatusSeeOther, "/import/preview?token="+token)
}

func (ih *ImportHandler) importPreviewHandler(c echo.Context) error {
	fromProtected, ok := c.Get("FROMPROTECTED").(bool)
	if !ok {
		return errors.New("invalid type for key 'FROMPROTECTED'")
	}

	token := c.QueryParam("token")
	candidates, ok := getCandidates(token)
	if !ok {
		setFlashmessages(c, "error", "Import session expired or not found. Please upload the file again.")
		return c.Redirect(http.StatusSeeOther, "/import")
	}

	previewView := auth_views.ImportPreview(token, candidates)
	c.Set("ISERROR", false)

	return renderView(c, auth_views.ImportPreviewIndex(
		"| Review import",
		"",
		fromProtected,
		c.Get("ISERROR").(bool),
		getFlashmessages(c, "error"),
		getFlashmessages(c, "success"),
		previewView,
	))
}

func (ih *ImportHandler) importConfirmHandler(c echo.Context) error {
	token := c.FormValue("token")
	candidates, ok := getCandidates(token)
	if !ok {
		setFlashmessages(c, "error", "Import session expired or not found. Please upload the file again.")
		return c.Redirect(http.StatusSeeOther, "/import")
	}

	userID, ok := c.Get(user_id_key).(int64)
	if !ok || userID == 0 {
		return echo.NewHTTPError(echo.ErrUnauthorized.Code, "Please provide valid credentials")
	}

	selected := c.Request().Form["selected"]
	toImport, err := filterSelected(candidates, selected)
	if err != nil {
		setFlashmessages(c, "error", err.Error())
		return c.Redirect(http.StatusSeeOther, "/import/preview?token="+token)
	}

	if len(toImport) == 0 {
		setFlashmessages(c, "error", "No items selected")
		return c.Redirect(http.StatusSeeOther, "/import/preview?token="+token)
	}

	if err := ih.locations.ImportLocations(c.Request().Context(), userID, toImport); err != nil {
		setFlashmessages(c, "error", fmt.Sprintf("Import failed: %s", err))
		return c.Redirect(http.StatusSeeOther, "/import/preview?token="+token)
	}

	importStore.Delete(token)

	setFlashmessages(c, "success", fmt.Sprintf("Imported %d location(s)", len(toImport)))
	return c.Redirect(http.StatusSeeOther, "/")
}

func newImportToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func getCandidates(token string) ([]services.ImportCandidate, bool) {
	if token == "" {
		return nil, false
	}
	raw, found := importStore.Get(token)
	if !found {
		return nil, false
	}
	candidates, ok := raw.([]services.ImportCandidate)
	return candidates, ok
}

func filterSelected(candidates []services.ImportCandidate, selected []string) ([]services.ImportCandidate, error) {
	byIndex := make(map[int]services.ImportCandidate, len(candidates))
	for _, c := range candidates {
		byIndex[c.Index] = c
	}

	var result []services.ImportCandidate
	for _, s := range selected {
		idx, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("invalid selection: %q", s)
		}
		if item, ok := byIndex[idx]; ok {
			result = append(result, item)
		}
	}

	return result, nil
}
