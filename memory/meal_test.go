package memory_test

import (
	"testing"

	"github.com/jameswhoughton/meals"
	"github.com/jameswhoughton/meals/contracts"
	"github.com/jameswhoughton/meals/memory"
)

func TestMemoryMealRepository(t *testing.T) {
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

func TestMemoryMealMetaDataRepository(t *testing.T) {
	init := func(meals []meals.Meal) (meals.MealMetaDataRepository, func()) {
		repo := &memory.MealMetaDataRepository{
			Store: meals,
		}

		return repo, func() {}
	}

	contract := contracts.MealMetaDataRepository{
		Repo: init,
	}

	contract.Test(t)

}
