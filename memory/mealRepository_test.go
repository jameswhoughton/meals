package memory_test

import (
	"testing"

	"github.com/jameswhoughton/meals"
	"github.com/jameswhoughton/meals/contracts"
	"github.com/jameswhoughton/meals/memory"
)

func TestMemoryRepository(t *testing.T) {
	init := func() (meals.MealRepository, func(id int), func()) {
		return &memory.MealRepository{
			Store: []meals.Meal{},
		}, func(_ int) {}, func() {}
	}

	contract := contracts.MealRepository{
		Repo: init,
	}

	contract.Test(t)

}
