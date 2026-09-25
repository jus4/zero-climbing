package services

import (
	"context"

	"github.com/jus4/zero-climbing/db/sqlc"
	"golang.org/x/crypto/bcrypt"
)

func NewUserServices(queries *sqlc.Queries) *UserServices {
	return &UserServices{
		queries: queries,
	}
}

type User struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Alias    string `json:"alias"`
	Password string `json:"password"`
}

type UserServices struct {
	queries *sqlc.Queries
}

func (us *UserServices) CreateUser(u User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), 8)
	if err != nil {
		return err
	}

	_, err = us.queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		Email:    u.Email,
		Alias:    u.Alias,
		Password: string(hashedPassword),
	})

	return err
}

func (us *UserServices) CheckEmail(email string) (User, error) {
	sqlcUser, err := us.queries.GetUserByEmail(context.Background(), email)
	if err != nil {
		return User{}, err
	}

	return User{
		ID:       sqlcUser.ID,
		Email:    sqlcUser.Email,
		Alias:    sqlcUser.Alias,
		Password: sqlcUser.Password,
	}, nil
}
