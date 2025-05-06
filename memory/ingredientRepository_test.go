package memory_test

import (
	"testing"

	"github.com/jameswhoughton/meals/internal/meals"
	"github.com/jameswhoughton/meals/memory"
)

func TestMemoryIngredientRepository(t *testing.T) {
	init := func(ingredients []meals.Ingredient) (meals.IngredientRepository, func()) {
		return &memory.IngredientRepository{
			Store: ingredients,
		}, func() {}
	}

	contract := meals.IngredientRepositoryContract{
		Repo: init,
	}

	contract.Test(t)

}
