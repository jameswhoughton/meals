package memory_test

import (
	"testing"

	"github.com/jameswhoughton/meals/internal/account"
	"github.com/jameswhoughton/meals/memory"
)

func TestAccountRepositoryContract(t *testing.T) {
	init := func() (account.Repository, func()) {
		return &memory.AccountRepository{
			Store: []account.User{},
		}, func() {}
	}

	contract := account.RepositoryContract{
		Repo: init,
	}

	contract.Test(t)
}
