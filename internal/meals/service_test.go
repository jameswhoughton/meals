package meals_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jameswhoughton/meals/internal/meals"
	"github.com/jameswhoughton/meals/memory"
)

func TestValidationErrorsWhenCreatingAMeal(t *testing.T) {
	mealRepository := memory.MealRepository{}
	tagRepository := memory.TagRepository{}
	service := meals.NewService(&mealRepository, &tagRepository)

	ctx := context.Background()

	type testCase struct {
		name           string
		meal           meals.Meal
		expectedErrors []string
	}

	testCases := []testCase{
		{
			name:           "Empty meal",
			meal:           meals.Meal{},
			expectedErrors: []string{"Name"},
		},
		{
			name: "Ingredient with zero quantity",
			meal: meals.Meal{
				Name: "test",
				Ingredients: []meals.Ingredient{
					{
						Id:       43,
						Name:     "ingredient 1",
						Quantity: 0,
					},
					{
						Name:     "ingredient 2",
						Quantity: 20,
					},
				},
			},
			expectedErrors: []string{"Ingredients.0"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := service.CreateMeal(ctx, &testCase.meal)

			if err == nil {
				t.Error("Expected validation error got none")
			}

			if !errors.Is(err, meals.ErrorFormInvalid{}) {
				t.Errorf("Expected error of type %T, got %T (%v)", meals.ErrorFormInvalid{}, err, err)
			}

			if len(testCase.meal.Errors) != len(testCase.expectedErrors) {
				t.Errorf("Expected %d validation errors, got %d", len(testCase.expectedErrors), len(testCase.meal.Errors))
			}

			for _, expectedError := range testCase.expectedErrors {
				if _, ok := testCase.meal.Errors[expectedError]; !ok {
					t.Errorf("Expected validation error %s missing", expectedError)
				}
			}
		})
	}

}

func TestServiceCanCreateAMeal(t *testing.T) {
	mealRepository := memory.MealRepository{}
	tagRepository := memory.TagRepository{}
	service := meals.NewService(&mealRepository, &tagRepository)

	ctx := context.Background()

	mealToCreate := meals.Meal{
		Name:  "New meal",
		Notes: "Something exciting",
		Tags: []meals.Tag{
			{
				Id:   1,
				Name: "Quick",
			},
		},
		Ingredients: []meals.Ingredient{
			{
				Id:       1,
				Name:     "Cheese",
				Quantity: 25,
				Unit:     "g",
			},
		},
	}

	createdMeal, err := service.CreateMeal(ctx, &mealToCreate)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(mealToCreate.Errors) > 0 {
		t.Errorf("Unexpected validation errors: %v", mealToCreate.Errors)
	}

	if createdMeal.Id == 0 {
		t.Error("Meal Id not set")
	}

	if len(mealRepository.Store) != 1 {
		t.Errorf("Expected 1 meal in the store, found %d", len(mealRepository.Store))
	}

	if mealRepository.Store[0].Name != mealToCreate.Name {
		t.Errorf("Expected meal in store to have name %s, found %s", mealToCreate.Name, mealRepository.Store[0].Name)
	}
}

func TestValidationErrorsWhenUpdatingAMeal(t *testing.T) {
	mealRepository := memory.MealRepository{}
	tagRepository := memory.TagRepository{}
	service := meals.NewService(&mealRepository, &tagRepository)

	ctx := context.Background()

	type testCase struct {
		name           string
		meal           meals.Meal
		expectedErrors []string
	}

	testCases := []testCase{
		{
			name:           "Empty meal",
			meal:           meals.Meal{},
			expectedErrors: []string{"Name"},
		},
		{
			name: "Ingredient with zero quantity",
			meal: meals.Meal{
				Name: "test",
				Ingredients: []meals.Ingredient{
					{
						Id:       43,
						Name:     "ingredient 1",
						Quantity: 0,
					},
					{
						Name:     "ingredient 2",
						Quantity: 20,
					},
				},
			},
			expectedErrors: []string{"Ingredients.0"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := service.UpdateMeal(ctx, &testCase.meal)

			if err == nil {
				t.Error("Expected validation error got none")
			}

			if !errors.Is(err, meals.ErrorFormInvalid{}) {
				t.Errorf("Expected error of type %T, got %T (%v)", meals.ErrorFormInvalid{}, err, err)
			}

			if len(testCase.meal.Errors) != len(testCase.expectedErrors) {
				t.Errorf("Expected %d validation errors, got %d", len(testCase.expectedErrors), len(testCase.meal.Errors))
			}

			for _, expectedError := range testCase.expectedErrors {
				if _, ok := testCase.meal.Errors[expectedError]; !ok {
					t.Errorf("Expected validation error %s missing", expectedError)
				}
			}
		})
	}

}
func TestServiceCanUpdateAMeal(t *testing.T) {
	mealRepository := memory.MealRepository{
		Store: []meals.Meal{
			{
				Id:   23,
				Name: "Old name",
			},
		},
	}
	tagRepository := memory.TagRepository{}
	service := meals.NewService(&mealRepository, &tagRepository)

	ctx := context.Background()

	mealToUpdate := meals.Meal{
		Id:   23,
		Name: "New name",
		Tags: []meals.Tag{
			{
				Id:   1,
				Name: "Quick",
			},
		},
		Ingredients: []meals.Ingredient{
			{
				Id:       1,
				Name:     "Cheese",
				Quantity: 25,
				Unit:     "g",
			},
		},
	}

	err := service.UpdateMeal(ctx, &mealToUpdate)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(mealToUpdate.Errors) > 0 {
		t.Errorf("Unexpected validation errors: %v", mealToUpdate.Errors)
	}

	if len(mealRepository.Store) != 1 {
		t.Errorf("Expected 1 meal in the store, found %d", len(mealRepository.Store))
	}

	if mealRepository.Store[0].Name != mealToUpdate.Name {
		t.Errorf("Name not updated in store: expected %s, found %s", mealToUpdate.Name, mealRepository.Store[0].Name)
	}
}
