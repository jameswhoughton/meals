package web

import (
	"errors"
	"testing"

	"github.com/jameswhoughton/meals"
)

type testUserService struct {
	users []meals.User
}

func (us *testUserService) Get(id int) (meals.User, error) {

	return meals.User{}, errors.New("user not found")
}
func (us *testUserService) GetFromSessionId(sessionId string) (meals.User, error)
func (us *testUserService) GetFromCredentials(email, password string) (meals.User, error)
func (us *testUserService) GetFromEmail(email string) (meals.User, error)
func (us *testUserService) Add(user meals.User) (meals.User, error)
func (us *testUserService) UpdateEmail(user meals.User, email string) error
func (us *testUserService) UpdatePassword(user meals.User, password string) error

func test_PostLoginHandler(t *testing.T) {

	t.Run("Returns error if email has been used", func(t *testing.T) {})

	t.Run("Returns error if the password and confirmation do not match", func(t *testing.T) {})

	t.Run("Creates a user and returns correctly", func(t *testing.T) {})
}
