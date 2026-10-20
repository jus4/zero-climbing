package handlers

import (
	"testing"

	"github.com/jus4/zero-climbing/services"
)

func TestSelectCandidates(t *testing.T) {
	candidates := []services.ImportCandidate{
		{Index: 0, Title: "a"},
		{Index: 1, Title: "b"},
		{Index: 2, Title: "c"},
	}

	got := selectCandidates(candidates, []int{2, 0, 0, 9})
	if len(got) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(got))
	}
	if got[0].Title != "c" || got[1].Title != "a" {
		t.Fatalf("unexpected order/content: %+v", got)
	}
}

func TestSelectCandidatesEmpty(t *testing.T) {
	if got := selectCandidates(nil, []int{0}); len(got) != 0 {
		t.Fatalf("expected no candidates, got %d", len(got))
	}
}
