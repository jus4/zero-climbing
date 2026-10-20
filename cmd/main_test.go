package main

import (
	"net/http"
	"testing"
)

func TestParseSameSite(t *testing.T) {
	cases := map[string]http.SameSite{
		"":       http.SameSiteLaxMode,
		"lax":    http.SameSiteLaxMode,
		"Lax":    http.SameSiteLaxMode,
		"Strict": http.SameSiteStrictMode,
		" none ": http.SameSiteNoneMode,
		"bogus":  http.SameSiteLaxMode,
	}

	for input, want := range cases {
		if got := parseSameSite(input); got != want {
			t.Errorf("parseSameSite(%q) = %v, want %v", input, got, want)
		}
	}
}
