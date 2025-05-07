package memory_test

import (
	"testing"

	"github.com/jameswhoughton/meals/memory"
	"github.com/jameswhoughton/meals/web"
)

func TestMemorySessionService(t *testing.T) {
	init := func() (web.SessionRepository, func()) {
		return &memory.SessionRepository{
			Store: []web.Session{},
		}, func() {}
	}

	contract := web.SessionRepositoryContract{
		Repo: init,
	}

	contract.Test(t)

}
