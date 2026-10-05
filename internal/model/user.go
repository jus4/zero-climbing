package model

import (
	"context"
	"time"

	"github.com/jus4/zero-climbing/db/sqlc"
)

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Alias     string    `json:"alias"`
	CreatedAt time.Time `json:"created_at"`
}

type UserModel struct {
	queries *sqlc.Queries
}

func NewUserModel(queries *sqlc.Queries) *UserModel {
	return &UserModel{
		queries: queries,
	}
}

func (m *UserModel) GetByEmail(ctx context.Context, email string) (User, error) {
	row, err := m.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return User{}, err
	}

	return toUser(row), nil
}

func (m *UserModel) GetByAlias(ctx context.Context, alias string) (User, error) {
	row, err := m.queries.GetUserByAlias(ctx, alias)
	if err != nil {
		return User{}, err
	}

	return toUser(row), nil
}

func (m *UserModel) GetByID(ctx context.Context, id int64) (User, error) {
	row, err := m.queries.GetUserByID(ctx, id)
	if err != nil {
		return User{}, err
	}

	return toUser(row), nil
}

func (m *UserModel) Create(ctx context.Context, email, alias string) (User, error) {
	row, err := m.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email: email,
		Alias: alias,
	})
	if err != nil {
		return User{}, err
	}

	return toUser(row), nil
}

func toUser(row sqlc.User) User {
	return User{
		ID:        row.ID,
		Email:     row.Email,
		Alias:     row.Alias,
		CreatedAt: row.CreatedAt.Time,
	}
}
