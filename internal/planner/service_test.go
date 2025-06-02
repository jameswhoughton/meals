package planner_test

import (
	"context"
	"testing"
	"time"

	"github.com/jameswhoughton/meals/internal/planner"
	"github.com/jameswhoughton/meals/memory"
)

func TestGetIngredientsValidation(t *testing.T) {
	repo := memory.NewPlannerRepository()

	service := planner.NewService(repo)

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
			expectedError: planner.ErrValidation,
		},
		{
			name:          "start date differs from end date by more than 4 weeks",
			startDate:     time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC),
			endDate:       time.Date(2025, 3, 5, 0, 0, 0, 0, time.UTC),
			expectedError: planner.ErrValidation,
		},
		{
			name:          "valid case",
			startDate:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:       time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		_, err := service.GetIngredients(context.Background(), testCase.startDate, testCase.endDate, 1)

		if err != testCase.expectedError {
			t.Errorf("Expected error %v, got %v", testCase.expectedError, err)
		}
	}
}
