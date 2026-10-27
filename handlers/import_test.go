package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/sessions"
	"github.com/jus4/zero-climbing/services"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func TestSelectCandidates(t *testing.T) {
	candidates := []services.ImportCandidate{
		{Index: 0, Title: "a"},
		{Index: 1, Title: "b"},
		{Index: 2, Title: "c"},
	}

	got := selectCandidates(candidates, []importSelection{
		{Index: 2, IsPublic: true},
		{Index: 0},
		{Index: 0, IsPublic: true},
		{Index: 9},
	})
	if len(got) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(got))
	}
	if got[0].Title != "c" || got[1].Title != "a" {
		t.Fatalf("unexpected order/content: %+v", got)
	}
	if !got[0].IsPublic {
		t.Fatalf("expected selected public location to be public")
	}
	if got[1].IsPublic {
		t.Fatalf("expected selected private location to be private")
	}
}

func TestSelectCandidatesEmpty(t *testing.T) {
	if got := selectCandidates(nil, []importSelection{{Index: 0}}); len(got) != 0 {
		t.Fatalf("expected no candidates, got %d", len(got))
	}
}

type fakeImportService struct {
	imported int
	skipped  int
	gotItems int
}

func (f *fakeImportService) ImportLocations(_ context.Context, _ int64, items []services.ImportCandidate) (int, int, error) {
	f.gotItems = len(items)
	return f.imported, f.skipped, nil
}

func TestImportConfirmReportsSkipped(t *testing.T) {
	const token = "test-import-token"
	importStore.Set(token, []services.ImportCandidate{
		{Index: 0, Title: "a"},
		{Index: 1, Title: "b"},
	}, time.Minute)
	defer importStore.Delete(token)

	imports := &fakeImportService{imported: 1, skipped: 1}
	api := NewAPIHandler(nil, nil, imports, nil, "")

	e := echo.New()
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("test-secret"))))
	e.GET("/session", func(c echo.Context) error {
		sess, _ := session.Get(auth_sessions_key, c)
		sess.Values = map[interface{}]interface{}{
			auth_key:    true,
			user_id_key: int64(10),
			alias_key:   "tester",
		}
		return sess.Save(c.Request(), c.Response())
	})
	e.POST("/confirm", api.ImportConfirm)

	loginRequest := httptest.NewRequest(http.MethodGet, "/session", nil)
	loginRecorder := httptest.NewRecorder()
	e.ServeHTTP(loginRecorder, loginRequest)
	cookies := loginRecorder.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected a session cookie")
	}

	body := `{"token":"` + token + `","selection":[{"index":0},{"index":1}]}`
	request := httptest.NewRequest(http.MethodPost, "/confirm", strings.NewReader(body))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Imported int `json:"imported"`
		Skipped  int `json:"skipped"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Imported != 1 || response.Skipped != 1 {
		t.Fatalf("response = %+v, want imported=1 skipped=1", response)
	}
	if imports.gotItems != 2 {
		t.Fatalf("service received %d items, want 2", imports.gotItems)
	}
}
