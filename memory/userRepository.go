package memory

import (
	"fmt"

	"github.com/jameswhoughton/meals/internal/auth"
)

type UserRepository struct {
	Store []auth.User
}

func (us *UserRepository) Get(filter auth.UserGet) (auth.User, error) {
	var (
		id    int
		email string
	)

	err := filter.Validate()

	if err != nil {
		return auth.User{}, fmt.Errorf("UserRepo.Get invalid parameter: %w", err)
	}

	if v := filter.Id; v != nil {
		id = *v
	}

	if v := filter.Email; v != nil {
		email = *v
	}

	for _, user := range us.Store {
		if user.Id == id || user.Email == email {
			return user, nil
		}
	}

	return auth.User{}, auth.ErrorUserNotFound{}
}

func (us *UserRepository) Create(user auth.User) (auth.User, error) {
	user.Id = len(us.Store) + 1

	us.Store = append(us.Store, user)

	return user, nil
}

func (us *UserRepository) Update(id int, form auth.UserUpdate) error {

	var index int
	found := false

	for i, user := range us.Store {
		if user.Id == id {
			index = i
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("Cannot find user to update: %w", auth.ErrorUserNotFound{})
	}

	if v := form.Email; v != nil {
		us.Store[index].Email = *v
	}

	if v := form.Name; v != nil {
		us.Store[index].Name = *v
	}

	if v := form.Password; v != nil {
		us.Store[index].Password = *v
	}

	return nil
}
