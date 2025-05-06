package auth

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

type UserRepositoryContract struct {
	Repo func() (UserRepository, func())
}

func (c *UserRepositoryContract) Test(t *testing.T) {
	t.Run("Can add, update and retrieve a user", func(t *testing.T) {
		repo, closeDown := c.Repo()
		defer closeDown()

		newUserEmail := "john@example.com"
		newUserPassword := "password123!"
		newName := "John Smith"

		// Add a new user
		form := User{
			Email:    newUserEmail,
			Password: newUserPassword,
			Name:     newName,
		}
		user, err := repo.Create(form)

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
		fetchedUsers, err := repo.Get(UserGet{Id: &user.Id})

		if err != nil {
			t.Error(err)
		}

		if fetchedUsers.Id != user.Id {
			t.Errorf("Expected ID %d found %d", user.Id, fetchedUsers.Id)
		}

		// Fetch an existing user by email
		fetchedUser, err := repo.Get(UserGet{Email: &newUserEmail})

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
			Name:         newName,
			Email:        newEmail,
			MealStartDay: newStartDay,
			UpdatedAt:    time.Now(),
		}

		err = repo.Update(fetchedUser.Id, update)

		if err != nil {
			t.Error(err)
		}

		updatedUser, err := repo.Get(UserGet{Email: &newEmail})

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
	})

	t.Run("Returns expected error if the user does not exist", func(t *testing.T) {
		repo, closeDown := c.Repo()
		defer closeDown()

		id := 1

		_, err := repo.Get(UserGet{Id: &id})

		if err == nil {
			t.Errorf("Expected error got nil")
		}

		if !errors.Is(err, ErrorUserNotFound{}) {
			t.Errorf("Expected UserNotFoundError, got %T", err)
		}
	})

	t.Run("Returns expected error if no search parameters passed to Get", func(t *testing.T) {
		repo, closeDown := c.Repo()
		defer closeDown()

		_, err := repo.Get(UserGet{})

		if err == nil {
			t.Errorf("Expected error, got nil")
		}

		if !errors.Is(err, ErrorUserGetInvalid{}) {
			fmt.Printf("%v, %T", err, err)
			t.Errorf("Expected ErrorUserGetInvalid, got %T", err)
		}

	})
}
