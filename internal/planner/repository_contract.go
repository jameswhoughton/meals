package planner

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

type RepositoryContract struct {
	Repo func() (Repository, func(meal Meal), func())
}

func (i RepositoryContract) Test(t *testing.T) {

	t.Run("Can add a meal to a date, fetch it and remove it", func(t *testing.T) {
		meal1 := Meal{
			Id:     1,
			UserId: 1,
			Name:   "Chicken Curry",
		}
		meal2 := Meal{
			Id:     2,
			UserId: 2,
			Name:   "Lamb Curry",
		}
		repo, seeder, closeDown := i.Repo()
		defer closeDown()

		seeder(meal1)
		seeder(meal2)

		ctx := context.Background()

		day := time.Date(2023, 1, 1, 12, 30, 00, 00, time.UTC)

		err := repo.Add(ctx, day, 1)

		if err != nil {
			t.Errorf("Unexpected error adding meal to day: %v", err)
		}

		mealFromDay, err := repo.Get(ctx, day, 1)

		if err != nil {
			t.Errorf("Unexpected error fetching a meal for a day: %v", err)
		}

		// Add a second meal for another user to the same day
		repo.Add(ctx, day, 2)

		if !cmp.Equal(meal1, mealFromDay) {
			t.Errorf("Fetched meal does not match the expected meal\n Expected:%#v\nReceived:%#v", meal1, mealFromDay)
		}

		err = repo.Clear(ctx, day, 1)

		if err != nil {
			t.Errorf("Unexpected error clearing a meal from a day: %v", err)
		}

		check, err := repo.Get(ctx, day, 1)

		if err == nil {
			t.Error("Expected error fetching a meal for a day, received nil")
		}

		if check.Id != 0 {
			t.Errorf("Expected no meal, found: %s", check.Name)
		}

		check, _ = repo.Get(ctx, day, 2)

		// Assert that the second user's meal has not been deleted
		if !cmp.Equal(meal2, check) {
			t.Errorf("Fetched meal does not match the expected meal\n Expected:%#v\nReceived:%#v", meal2, check)
		}
	})

	t.Run("The time should be ignored when interacting with the planner", func(t *testing.T) {
		repo, seeder, closeDown := i.Repo()
		defer closeDown()

		seeder(Meal{
			Id:     23,
			UserId: 4,
			Name:   "Salmon stir-fry",
		})

		ctx := context.Background()

		day := time.Date(2025, 3, 20, 11, 30, 00, 00, time.UTC)
		laterOn := day.Add(3 * time.Hour)

		err := repo.Add(ctx, day, 23)

		if err != nil {
			t.Errorf("Unexpected error adding meal to day: %v", err)
		}

		fetchedMeal, err := repo.Get(ctx, laterOn, 4)

		if err != nil {
			t.Errorf("Unexpected error fetching a meal for a day: %v", err)
		}

		if fetchedMeal.Id != 23 {
			t.Errorf("Expected meal %d, found %d", 23, fetchedMeal.Id)
		}

	})
}
