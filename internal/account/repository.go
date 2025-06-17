package account

import (
	"context"
	"errors"
	"time"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	Id           int
	Name         string
	Email        string
	Password     string
	MealStartDay int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserUpdate struct {
	Id           int
	Name         string
	Email        string
	Password     *string
	MealStartDay int
	UpdatedAt    time.Time
}

type Repository interface {

	// Get a user that matches the given id
	//
	// Returns ErrUserNotFound if the id does not match any users
	GetById(ctx context.Context, id int) (User, error)

	// Get a user that matches the given email
	//
	// Returns ErrUserNotFound if the email does not match any users
	GetByEmail(ctx context.Context, email string) (User, error)

	// Create a new user
	//
	// Returns the new user if created successfully.
	Create(ctx context.Context, user User) (User, error)

	// Updates an existing user
	//
	// Returns ErrUserNotFound if the given user.Id does not exist.
	Update(ctx context.Context, user UserUpdate) error

	// Deletes an existing user
	//
	// Does nothing if the userId does not exist.
	Delete(ctx context.Context, userId int) error
}
