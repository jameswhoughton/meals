package memory_test

import (
	"testing"

	"github.com/jameswhoughton/meals/internal/auth"
	"github.com/jameswhoughton/meals/memory"
)

func TestMemoryUserService(t *testing.T) {
	init := func() (auth.UserRepository, func()) {
		return &memory.UserRepository{
			Store: []auth.User{},
		}, func() {}
	}

	contract := auth.UserRepositoryContract{
		Repo: init,
	}

	contract.Test(t)

}
