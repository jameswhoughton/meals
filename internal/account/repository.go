package account

import (
	"context"
	"time"
)

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

type ErrorUserNotFound struct {
}

func (e ErrorUserNotFound) Error() string {
	return "User not found"
}

type ErrorGetFormInvalid struct{}

func (e ErrorGetFormInvalid) Error() string {
	return "Expects at least 1 param, none provided"
}

type GetForm struct {
	Id    *int
	Email *string
}

type Repository interface {
	Get(ctx context.Context, form GetForm) (User, error)
	Create(ctx context.Context, user User) (User, error)
	Update(ctx context.Context, form UserUpdate) error
}
