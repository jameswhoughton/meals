package auth

import (
	"time"
)

type UserGet struct {
	Id    *int
	Email *string
}

type ErrorUserGetInvalid struct{}

func (e ErrorUserGetInvalid) Error() string {
	return "Expects at least 1 param, none provided"
}

func (ug UserGet) Validate() error {
	if ug.Id == nil && ug.Email == nil {
		return ErrorUserGetInvalid{}
	}

	return nil
}

type UserUpdate struct {
	Email        string
	Name         string
	Password     *string
	MealStartDay int
	UpdatedAt    time.Time
}

type ErrorUserNotFound struct {
}

func (e ErrorUserNotFound) Error() string {
	return "User not found"
}

type UserRepository interface {
	Get(filter UserGet) (User, error)
	Create(user User) (User, error)
	Update(id int, form UserUpdate) error
}
