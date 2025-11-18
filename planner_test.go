package meals_test

import (
	"context"
	"testing"
	"time"

	"github.com/jameswhoughton/meals"
	"github.com/jameswhoughton/meals/memory"
)

func TestGetIngredientsValidation(t *testing.T) {
	plannerRepo := memory.NewPlannerRepository()
	mealRepo := memory.MealRepository{}
	mealMetaDataRepo := memory.MealMetaDataRepository{}

	service := meals.NewPlannerService(plannerRepo, &mealRepo, &mealMetaDataRepo)

	type testCase struct {
		name          string
		startDate     time.Time
		endDate       time.Time
		expectedError error
	}

	testCases := []testCase{
		{
			name:          "start date after end date",
			startDate:     time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC),
			endDate:       time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
			expectedError: meals.ErrValidation,
		},
		{
			name:          "start date differs from end date by more than 4 weeks",
			startDate:     time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC),
			endDate:       time.Date(2025, 3, 5, 0, 0, 0, 0, time.UTC),
			expectedError: meals.ErrValidation,
		},
		{
			name:          "valid case",
			startDate:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:       time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := service.GetIngredients(context.Background(), testCase.startDate, testCase.endDate, 1)

			if err != testCase.expectedError {
				t.Errorf("Expected error %v, got %v", testCase.expectedError, err)
			}
		})
	}
}

func TestGetMealsCanReturnAGivenNumberOfDays(t *testing.T) {
	plannerRepo := memory.NewPlannerRepository()
	mealRepo := memory.MealRepository{}
	mealMetaDataRepo := memory.MealMetaDataRepository{}

	service := meals.NewPlannerService(plannerRepo, &mealRepo, &mealMetaDataRepo)

	type testCase struct {
		name         string
		startDate    time.Time
		numberOfDays int
		expectError  bool
	}

	testCases := []testCase{
		{
			name:         "single day",
			startDate:    time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC),
			numberOfDays: 1,
			expectError:  false,
		},
		{
			name:         "multiple days",
			startDate:    time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
			numberOfDays: 4,
			expectError:  false,
		},
		{
			name:         "zero",
			startDate:    time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC),
			numberOfDays: 0,
			expectError:  true,
		},
		{
			name:         "negative",
			startDate:    time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC),
			numberOfDays: -1,
			expectError:  true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			days, err := service.GetMeals(context.Background(), testCase.startDate, testCase.numberOfDays, 1)

			if testCase.expectError && err != nil {
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)

				return
			}

			if len(days) != testCase.numberOfDays {
				t.Errorf("expected %d days, got %d", testCase.numberOfDays, len(days))

				return
			}

			expectedDate := testCase.startDate

			for _, day := range days {
				if day.Date != expectedDate.Format("2006-01-02") {
					t.Errorf("expected date %s, got %s", expectedDate.Format("2006-01-02"), day.Date)
				}

				expectedDate = expectedDate.Add(time.Hour * 24)
			}
		})
	}
}
