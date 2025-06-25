package memory

import (
	"context"
	"fmt"

	"github.com/jameswhoughton/meals/internal/account"
)

type AccountRepository struct {
	store []account.User
}

func (ar *AccountRepository) GetById(ctx context.Context, id int) (account.User, error) {
	for _, user := range ar.store {
		if user.Id == id {
			return user, nil
		}
	}

	return account.User{}, account.ErrUserNotFound
}

func (ar *AccountRepository) GetByEmail(ctx context.Context, email string) (account.User, error) {
	for _, user := range ar.store {
		if user.Email == email {
			return user, nil
		}
	}

	return account.User{}, account.ErrUserNotFound
}

func (ar *AccountRepository) Create(ctx context.Context, user account.User) (account.User, error) {
	user.Id = len(ar.store) + 1

	ar.store = append(ar.store, user)

	return user, nil
}

func (ar *AccountRepository) Update(ctx context.Context, form account.UserUpdate) error {
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
		return fmt.Errorf("Cannot find user to update: %w", account.ErrUserNotFound)
	}

	ar.store[index].Email = form.Email

	ar.store[index].Name = form.Name

	ar.store[index].MealStartDay = form.MealStartDay

	if v := form.Password; v != nil {
		ar.store[index].Password = *v
	}

	return nil
}

func (ar *AccountRepository) Delete(ctx context.Context, userId int) error {
	var store []account.User

	for _, user := range ar.store {
		if user.Id == userId {
			continue
		}

		store = append(store, user)
	}

	ar.store = store

	return nil
}
