package memory_test

import (
	"testing"

	"github.com/jameswhoughton/meals/internal/planner"
	"github.com/jameswhoughton/meals/memory"
)

func TestMemoryPlannerRepository(t *testing.T) {
	init := func() (planner.Repository, func(m planner.Meal), func()) {
		store := make(map[string]int)

		repo := &memory.PlannerRepository{
			Planner: store,
		}

		return repo, func(m planner.Meal) { repo.Meals = append(repo.Meals, m) }, func() {}
	}

	contract := planner.RepositoryContract{
		Repo: init,
	}

	contract.Test(t)

}
