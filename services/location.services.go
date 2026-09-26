package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jus4/zero-climbing/db/sqlc"
)

func NewLocationServices(pool *pgxpool.Pool, queries *sqlc.Queries) *LocationServices {
	return &LocationServices{
		pool:    pool,
		queries: queries,
	}
}

type Location struct {
	ID          int64           `json:"id"`
	Geojson     json.RawMessage `json:"geojson"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Username    string          `json:"username"`
}

type LocationServices struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func (s *LocationServices) GetPointLocations(ctx context.Context) ([]Location, error) {
	rows, err := s.queries.GetPointLocations(ctx)
	if err != nil {
		return nil, fmt.Errorf("get point locations: %w", err)
	}

	locations := make([]Location, 0, len(rows))
	for _, row := range rows {
		locations = append(locations, Location{
			ID:          row.ID,
			Geojson:     row.Geojson,
			Title:       row.Title,
			Description: row.Description,
			Username:    row.Alias,
		})
	}

	return locations, nil
}

func (s *LocationServices) ImportLocations(ctx context.Context, userID int64, items []ImportCandidate) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	for _, item := range items {
		_, err := qtx.CreateLocation(ctx, sqlc.CreateLocationParams{
			UserID:      userID,
			Geojson:     item.Geojson,
			Title:       item.Title,
			Description: item.Description,
		})
		if err != nil {
			return fmt.Errorf("create location: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
