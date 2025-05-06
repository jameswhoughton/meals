package memory_test

import (
	"testing"

	"github.com/jameswhoughton/meals/internal/meals"
	"github.com/jameswhoughton/meals/memory"
)

func TestMemorytagRepository(t *testing.T) {
	init := func(tags []meals.Tag) (meals.TagRepository, func()) {
		return &memory.TagRepository{
			Store: tags,
		}, func() {}
	}

	contract := meals.TagRepositoryContract{
		Repo: init,
	}

	contract.Test(t)

}
