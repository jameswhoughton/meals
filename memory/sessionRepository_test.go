package memory_test

import (
	"testing"

	"github.com/jameswhoughton/meals/internal/auth"
	"github.com/jameswhoughton/meals/memory"
)

func TestMemorySessionService(t *testing.T) {
	init := func() (auth.SessionRepository, func()) {
		return &memory.SessionRepository{
			Store: []auth.Session{},
		}, func() {}
	}

	contract := auth.SessionRepositoryContract{
		Repo: init,
	}

	contract.Test(t)

}
