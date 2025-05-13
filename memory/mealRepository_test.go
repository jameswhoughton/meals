package memory_test

import (
	"testing"
	"time"

	"github.com/jameswhoughton/meals/internal/meals"
	"github.com/jameswhoughton/meals/memory"
)

func TestMemoryRepository(t *testing.T) {
	init := func() (meals.Repository, func(id int), func()) {
		return &memory.MealRepository{
			Store:    []meals.Meal{},
			Calendar: make(map[int][]time.Time),
		}, func(_ int) {}, func() {}
	}

	contract := meals.RepositoryContract{
		Repo: init,
	}

	contract.Test(t)

}
