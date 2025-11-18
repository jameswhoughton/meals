package memory_test

import (
	"testing"

	"github.com/jameswhoughton/meals"
	"github.com/jameswhoughton/meals/contracts"
	"github.com/jameswhoughton/meals/memory"
)

func TestMemoryPlannerRepository(t *testing.T) {
	init := func(testData []meals.Meal) (meals.PlannerRepository, func()) {
		store := make(map[string][]int)

		repo := &memory.PlannerRepository{
			Planner: store,
			Meals:   testData,
		}

		return repo, func() {}
	}

	contract := contracts.PlannerRepository{
		Repo: init,
	}

	contract.Test(t)

}
