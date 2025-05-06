package memory_test

import (
	"testing"

	"github.com/jameswhoughton/meals/internal/planner"
	"github.com/jameswhoughton/meals/memory"
)

func TestMemoryPlannerRepository(t *testing.T) {
	init := func(meals []planner.Meal) (planner.Repository, func()) {
		store := make(map[string]int)

		return &memory.PlannerRepository{
			Planner: store,
			Meals:   meals,
		}, func() {}
	}

	contract := planner.RepositoryContract{
		Repo: init,
	}

	contract.Test(t)

}
