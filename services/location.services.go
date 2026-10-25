package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
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
	UserID      int64           `json:"user_id"`
	Geojson     json.RawMessage `json:"geojson"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	IsPublic    bool            `json:"is_public"`
	Username    string          `json:"username"`
}

type LocationServices struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func (s *LocationServices) GetPublicPointLocations(ctx context.Context) ([]Location, error) {
	rows, err := s.queries.GetPublicPointLocations(ctx)
	if err != nil {
		return nil, fmt.Errorf("get public point locations: %w", err)
	}

	locations := make([]Location, 0, len(rows))
	for _, row := range rows {
		locations = append(locations, Location{
			ID:          row.ID,
			UserID:      row.UserID,
			Geojson:     row.Geojson,
			Title:       row.Title,
			Description: row.Description,
			IsPublic:    row.IsPublic,
			Username:    row.Alias,
		})
	}

	return locations, nil
}

func (s *LocationServices) GetPrivatePointLocations(ctx context.Context, userID int64) ([]Location, error) {
	rows, err := s.queries.GetPrivatePointLocationsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get private point locations: %w", err)
	}

	locations := make([]Location, 0, len(rows))
	for _, row := range rows {
		locations = append(locations, Location{
			ID:          row.ID,
			UserID:      row.UserID,
			Geojson:     row.Geojson,
			Title:       row.Title,
			Description: row.Description,
			IsPublic:    row.IsPublic,
			Username:    row.Alias,
		})
	}

	return locations, nil
}

func (s *LocationServices) UpdateLocation(ctx context.Context, id, userID int64, title, description string, isPublic bool) error {
	_, err := s.queries.UpdateLocation(ctx, sqlc.UpdateLocationParams{
		ID:          id,
		UserID:      userID,
		Title:       title,
		Description: description,
		IsPublic:    isPublic,
	})
	if err != nil {
		return fmt.Errorf("update location: %w", err)
	}
	return nil
}

func (s *LocationServices) DeleteLocation(ctx context.Context, id, userID int64) error {
	count, err := s.queries.DeleteLocation(ctx, sqlc.DeleteLocationParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("delete location: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("delete location: %w", pgx.ErrNoRows)
	}
	return nil
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
			IsPublic:    item.IsPublic,
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
