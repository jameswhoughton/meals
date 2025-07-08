package memory

import (
	"context"
	"fmt"

	"github.com/jameswhoughton/meals"
)

type UserRepository struct {
	store []meals.User
}

func (ar *UserRepository) GetById(ctx context.Context, id int) (meals.User, error) {
	for _, user := range ar.store {
		if user.Id == id {
			return user, nil
		}
	}

	return meals.User{}, meals.ErrUserNotFound
}

func (ar *UserRepository) GetByEmail(ctx context.Context, email string) (meals.User, error) {
	for _, user := range ar.store {
		if user.Email == email {
			return user, nil
		}
	}

	return meals.User{}, meals.ErrUserNotFound
}

func (ar *UserRepository) Create(ctx context.Context, user meals.User) (meals.User, error) {
	user.Id = len(ar.store) + 1

	ar.store = append(ar.store, user)

	return user, nil
}

func (ar *UserRepository) Update(ctx context.Context, form meals.UserUpdate) error {
	var index int
	found := false

	for i, user := range ar.store {
		if user.Id == form.Id {
			index = i
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("Cannot find user to update: %w", meals.ErrUserNotFound)
	}

	ar.store[index].Email = form.Email

	ar.store[index].Name = form.Name

	ar.store[index].MealStartDay = form.MealStartDay

	if v := form.Password; v != nil {
		ar.store[index].Password = *v
	}

	return nil
}

func (ar *UserRepository) Delete(ctx context.Context, userId int) error {
	var store []meals.User

	for _, user := range ar.store {
		if user.Id == userId {
			continue
		}

		store = append(store, user)
	}

	ar.store = store

	return nil
}
