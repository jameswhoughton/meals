package memory_test

import (
	"testing"

	"github.com/jameswhoughton/meals/internal/planner"
	"github.com/jameswhoughton/meals/memory"
)

func TestMemoryPlannerRepository(t *testing.T) {
	init := func() (planner.Repository, func(m planner.Meal, i []planner.Ingredient), func()) {
		store := make(map[string]int)
		ingredients := make(map[int][]planner.Ingredient)

		repo := &memory.PlannerRepository{
			Planner:     store,
			Ingredients: ingredients,
		}

		return repo, func(m planner.Meal, i []planner.Ingredient) {
			repo.Meals = append(repo.Meals, m)
			repo.Ingredients[m.Id] = i
		}, func() {}
	}

	contract := planner.RepositoryContract{
		Repo: init,
	}

	contract.Test(t)

}
