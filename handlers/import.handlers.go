package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jus4/zero-climbing/services"
	"github.com/patrickmn/go-cache"
)

var importStore = cache.New(30*time.Minute, 1*time.Hour)

const maxUploadSize = 50 << 20

type ImportService interface {
	ImportLocations(ctx context.Context, userID int64, items []services.ImportCandidate) error
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

type importSelection struct {
	Index    int  `json:"index"`
	IsPublic bool `json:"is_public"`
}

func selectCandidates(candidates []services.ImportCandidate, selection []importSelection) []services.ImportCandidate {
	byIndex := make(map[int]services.ImportCandidate, len(candidates))
	for _, c := range candidates {
		byIndex[c.Index] = c
	}

	seen := make(map[int]bool, len(selection))
	var result []services.ImportCandidate
	for _, sel := range selection {
		if seen[sel.Index] {
			continue
		}
		if item, ok := byIndex[sel.Index]; ok {
			seen[sel.Index] = true
			item.IsPublic = sel.IsPublic
			result = append(result, item)
		}
	}

	return result
}
