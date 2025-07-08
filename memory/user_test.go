package memory_test

import (
	"testing"

	"github.com/jameswhoughton/meals"
	"github.com/jameswhoughton/meals/contracts"
	"github.com/jameswhoughton/meals/memory"
)

func TestUserRepositoryContract(t *testing.T) {
	init := func() (meals.UserRepository, func()) {
		return &memory.UserRepository{}, func() {}
	}

	contract := contracts.UserRepository{
		Repo: init,
	}

	contract.Test(t)
}
