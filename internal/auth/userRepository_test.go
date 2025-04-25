package auth_test

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jameswhoughton/meals/database"
	"github.com/jameswhoughton/meals/internal/auth"
	"github.com/jameswhoughton/meals/memory"
)

type UserRepositoryContract struct {
	repo func() (auth.UserRepository, func())
}

func (c *UserRepositoryContract) Test(t *testing.T) {
	t.Run("Can add, update and retrieve a user", func(t *testing.T) {
		repo, closeDown := c.repo()
		defer closeDown()

		newUserEmail := "john@example.com"
		newUserPassword := "password123!"
		newName := "John Smith"

		// Add a new user
		form := auth.User{
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
		fetchedUsers, err := repo.Get(auth.UserGet{Id: &user.Id})

		if err != nil {
			t.Error(err)
		}

		if fetchedUsers.Id != user.Id {
			t.Errorf("Expected ID %d found %d", user.Id, fetchedUsers.Id)
		}

		// Fetch an existing user by email
		fetchedUser, err := repo.Get(auth.UserGet{Email: &newUserEmail})

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
		update := auth.UserUpdate{
			Name:         newName,
			Email:        newEmail,
			MealStartDay: newStartDay,
			UpdatedAt:    time.Now(),
		}

		err = repo.Update(fetchedUser.Id, update)

		if err != nil {
			t.Error(err)
		}

		updatedUser, err := repo.Get(auth.UserGet{Email: &newEmail})

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
		repo, closeDown := c.repo()
		defer closeDown()

		id := 1

		_, err := repo.Get(auth.UserGet{Id: &id})

		if err == nil {
			t.Errorf("Expected error got nil")
		}

		if !errors.Is(err, auth.ErrorUserNotFound{}) {
			t.Errorf("Expected UserNotFoundError, got %T", err)
		}
	})

	t.Run("Returns expected error if no search parameters passed to Get", func(t *testing.T) {
		repo, closeDown := c.repo()
		defer closeDown()

		_, err := repo.Get(auth.UserGet{})

		if err == nil {
			t.Errorf("Expected error, got nil")
		}

		if !errors.Is(err, auth.ErrorUserGetInvalid{}) {
			fmt.Printf("%v, %T", err, err)
			t.Errorf("Expected ErrorUserGetInvalid, got %T", err)
		}

	})
}

func TestDatabaseUserService(t *testing.T) {
	init := func() (auth.UserRepository, func()) {
		conn, err := sql.Open("sqlite3", "meals.db")

		if err != nil {
			log.Fatal(err)
		}

		err = database.Migrate(conn)

		if err != nil {
			log.Fatal(err)
		}

		closeDown := func() {
			os.Remove("meals.db")
		}
		return database.NewUserRespository(conn), closeDown
	}

	contract := UserRepositoryContract{
		init,
	}

	contract.Test(t)

}

func TestMemoryUserService(t *testing.T) {
	init := func() (auth.UserRepository, func()) {
		return &memory.UserRepository{
			Store: []auth.User{},
		}, func() {}
	}

	contract := UserRepositoryContract{
		init,
	}

	contract.Test(t)

}
