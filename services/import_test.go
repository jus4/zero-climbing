package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseImportFileGPX(t *testing.T) {
	content := `<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1" creator="test">
  <trk>
    <name>Track</name>
    <trkseg>
      <trkpt lat="60.1" lon="24.9"/>
      <trkpt lat="60.2" lon="24.95"/>
    </trkseg>
  </trk>
</gpx>`

	path := filepath.Join(t.TempDir(), "test.gpx")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	candidates, err := ParseImportFile(path, "test.gpx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}
	if candidates[0].Type != "line" || candidates[0].Title != "Track" {
		t.Fatalf("unexpected candidate: %+v", candidates[0])
	}
}

func TestParseImportFileUnsupported(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(path, []byte("not a map file"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := ParseImportFile(path, "test.txt"); err == nil {
		t.Fatal("expected an error for an unsupported format")
	}
}
