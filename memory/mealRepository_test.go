package memory_test

import (
	"testing"

	"github.com/jameswhoughton/meals/internal/meals"
	"github.com/jameswhoughton/meals/memory"
)

func TestMemoryRepository(t *testing.T) {
	init := func() (meals.Repository, func(id int), func()) {
		return &memory.MealRepository{
			Store: []meals.Meal{},
		}, func(_ int) {}, func() {}
	}

	contract := meals.RepositoryContract{
		Repo: init,
	}

	contract.Test(t)

}
