package contracts

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jameswhoughton/meals"
)

type PlannerRepository struct {
	Repo func(testData []meals.Meal) (meals.PlannerRepository, func())
}

func (i PlannerRepository) Test(t *testing.T) {

	t.Run("Can add meals to a date, fetch them and remove them", func(t *testing.T) {
		testData := []meals.Meal{
			{
				Id:     1,
				UserId: 1,
				Name:   "Chicken Curry",
			},
			{
				Id:     3,
				UserId: 1,
				Name:   "Vegetable Curry",
			},
			{
				Id:     2,
				UserId: 2,
				Name:   "Lamb Curry",
			},
		}
		repo, closeDown := i.Repo(testData)
		defer closeDown()

		ctx := context.Background()

		day := time.Date(2023, 1, 1, 12, 30, 00, 00, time.UTC)

		err := repo.Add(ctx, day, 1)

		if err != nil {
			t.Errorf("Unexpected error adding meal to day: %v", err)
		}

		err = repo.Add(ctx, day, 3)

		if err != nil {
			t.Errorf("Unexpected error adding meal to day: %v", err)
		}

		// Add a meal for another user to the same day
		err = repo.Add(ctx, day, 2)

		if err != nil {
			t.Errorf("Unexpected error adding meal to day: %v", err)
		}

		days, err := repo.GetMealIdsInRange(ctx, day, day, 1)

		if err != nil {
			t.Errorf("Unexpected error fetching a meal for a day: %v", err)
		}
		fmt.Println(days)

		if len(days[meals.GetDateKey(day)]) != 2 {
			t.Errorf("expected 2 meals, got %d", len(days[meals.GetDateKey(day)]))
		}

		if testData[0].Id != days[meals.GetDateKey(day)][0] {
			t.Errorf("Fetched meal does not match the expected meal\n Expected:%d\nReceived:%d", testData[0].Id, days[meals.GetDateKey(day)])
		}

		if testData[1].Id != days[meals.GetDateKey(day)][1] {
			t.Errorf("Fetched meal does not match the expected meal\n Expected:%d\nReceived:%d", testData[0].Id, days[meals.GetDateKey(day)])
		}

		err = repo.Clear(ctx, day, 1)

		if err != nil {
			t.Errorf("Unexpected error clearing a meal from a day: %v", err)
		}

		check, err := repo.GetMealIdsInRange(ctx, day, day, 1)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(check[meals.GetDateKey(day)]) != 0 {
			t.Errorf("Expected no meals, found: %d", len(check[meals.GetDateKey(day)]))
		}

		check, _ = repo.GetMealIdsInRange(ctx, day, day, 2)

		// Assert that the second user's meal has not been deleted
		if testData[2].Id != check[meals.GetDateKey(day)][0] {
			t.Errorf("Fetched meal does not match the expected meal\n Expected:%#v\nReceived:%#v", testData[1], check)
		}
	})

	t.Run("The time should be ignored when interacting with the planner", func(t *testing.T) {
		testData := []meals.Meal{
			{
				Id:     23,
				UserId: 4,
				Name:   "Salmon stir-fry",
			},
		}
		repo, closeDown := i.Repo(testData)
		defer closeDown()

		ctx := context.Background()

		day := time.Date(2025, 3, 20, 11, 30, 00, 00, time.UTC)
		laterOn := day.Add(3 * time.Hour)

		err := repo.Add(ctx, day, 23)

		if err != nil {
			t.Errorf("Unexpected error adding meal to day: %v", err)
		}

		fetchedMeal, err := repo.GetMealIdsInRange(ctx, day, laterOn, 4)

		if err != nil {
			t.Errorf("Unexpected error fetching a meal for a day: %v", err)
		}

		if fetchedMeal[meals.GetDateKey(day)][0] != 23 {
			t.Errorf("Expected meal %d, found %d", 23, fetchedMeal[meals.GetDateKey(day)])
		}

	})

	t.Run("Returns empty slice if no meal is set", func(t *testing.T) {
		repo, closeDown := i.Repo(nil)
		defer closeDown()

		startDate := time.Now()
		endDate := startDate.AddDate(0, 0, 7)

		days, err := repo.GetMealIdsInRange(context.Background(), startDate, endDate, 0)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		for _, meals := range days {
			if len(meals) != 0 {
				t.Errorf("expected no meals, got %d", meals)
			}
		}
	})
}
