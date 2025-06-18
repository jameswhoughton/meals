package account

import (
	"context"
	"errors"
	"testing"
	"time"
)

type RepositoryContract struct {
	Repo func() (Repository, func())
}

func (c *RepositoryContract) Test(t *testing.T) {
	t.Run("Can add, update, retrieve and delete a user", func(t *testing.T) {
		repo, closeDown := c.Repo()
		defer closeDown()

		ctx := context.Background()

		newUserEmail := "john@example.com"
		newUserPassword := "password123!"
		newName := "John Smith"

		// Add a new user
		form := User{
			Email:     newUserEmail,
			Password:  newUserPassword,
			Name:      newName,
			CreatedAt: time.Now(),
		}

		form.UpdatedAt = form.CreatedAt

		user, err := repo.Create(ctx, form)

		if err != nil {
			t.Error(err)
		}

		if user.Id == 0 {
			t.Errorf("Expected incremented ID found 0")
		}

		if user.Email != newUserEmail {
			t.Errorf("Expected email %s found %s", newUserEmail, user.Email)
		}

		if user.Name != newName {
			t.Errorf("Expected name %s found %s", newName, user.Name)
		}

		// Fetch an existing user by ID
		fetchedUsers, err := repo.GetById(ctx, user.Id)

		if err != nil {
			t.Error(err)
		}

		if fetchedUsers.Id != user.Id {
			t.Errorf("Expected ID %d found %d", user.Id, fetchedUsers.Id)
		}

		// Fetch an existing user by email
		fetchedUser, err := repo.GetByEmail(ctx, user.Email)

		if err != nil {
			t.Error(err)
		}

		if fetchedUser.Id != user.Id {
			t.Errorf("Expected ID %d found %d", user.Id, fetchedUser.Id)
		}

		if fetchedUser.Password == "" {
			t.Error("Password missing")
		}

		newName = "James Smith"
		newEmail := "james.smith@example.com"
		newStartDay := 3

		// Update User
		update := UserUpdate{
			Id:           fetchedUser.Id,
			Name:         newName,
			Email:        newEmail,
			MealStartDay: newStartDay,
			UpdatedAt:    time.Now(),
		}

		err = repo.Update(ctx, update)

		if err != nil {
			t.Error(err)
		}

		updatedUser, err := repo.GetByEmail(ctx, newEmail)

		if err != nil {
			t.Error(err)
		}

		if user.Id != updatedUser.Id {
			t.Errorf("Expected ID %d found %d", user.Id, updatedUser.Id)
		}

		if newEmail != updatedUser.Email {
			t.Errorf("Expected email %s found %s", newEmail, updatedUser.Email)
		}

		if newName != updatedUser.Name {
			t.Errorf("Expected name %s found %s", newName, updatedUser.Name)
		}

		if newStartDay != updatedUser.MealStartDay {
			t.Errorf("Expected start day %d found %d", newStartDay, updatedUser.MealStartDay)
		}

		err = repo.Delete(ctx, updatedUser.Id)

		if err != nil {
			t.Errorf("Unexpected error deleting user: %v", err)
		}

		_, err = repo.GetById(ctx, update.Id)

		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("Expected user not found error, got %T: %v", err, err)
		}
	})

	t.Run("Returns expected error if the user does not exist", func(t *testing.T) {
		repo, closeDown := c.Repo()
		defer closeDown()

		ctx := context.Background()

		_, err := repo.GetById(ctx, 1)

		if err == nil {
			t.Errorf("Expected error got nil")
		}

		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("Expected UserNotFoundError, got %T", err)
		}

		_, err = repo.GetByEmail(ctx, "test@example.com")

		if err == nil {
			t.Errorf("Expected error got nil")
		}

		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("Expected UserNotFoundError, got %T", err)
		}

		err = repo.Update(ctx, UserUpdate{Id: 1})

		if err == nil {
			t.Errorf("Expected error got nil")
		}

		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("Expected UserNotFoundError, got %T", err)
		}
	})
}
