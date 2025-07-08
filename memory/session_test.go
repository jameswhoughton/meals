package memory_test

import (
	"testing"

	"github.com/jameswhoughton/meals/contracts"
	"github.com/jameswhoughton/meals/memory"
	"github.com/jameswhoughton/meals/web"
)

func TestMemorySessionService(t *testing.T) {
	init := func() (web.SessionRepository, func(id int), func()) {
		return &memory.SessionRepository{}, func(_ int) {}, func() {}
	}

	contract := contracts.SessionRepository{
		Repo: init,
	}

	contract.Test(t)

}
