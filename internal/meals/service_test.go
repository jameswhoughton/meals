package meals_test

import (
	"errors"
	"testing"

	"github.com/jameswhoughton/meals/internal/meals"
	"github.com/jameswhoughton/meals/memory"
)

func TestValidationErrorsWhenCreatingAMeal(t *testing.T) {
	mealRepository := memory.MealRepository{}
	ingredientRepository := memory.IngredientRepository{}

	service := meals.NewService(&mealRepository, &ingredientRepository)

	type testCase struct {
		name           string
		meal           meals.Meal
		expectedErrors []string
	}

	testCases := []testCase{
		{
			name:           "Empty meal",
			meal:           meals.Meal{},
			expectedErrors: []string{"Name", "Ingredients"},
		},
		{
			name: "Ingredient with zero quantity",
			meal: meals.Meal{
				Name: "test",
				Ingredients: []meals.MealIngredient{
					{
						Id:       43,
						Name:     "ingredient 1",
						Quantity: 0,
					},
					{
						Name:     "ingredient 2",
						Quantity: 20,
						IsMain:   true,
					},
				},
			},
			expectedErrors: []string{"Ingredients.0"},
		},
		{
			name: "No main ingredient",
			meal: meals.Meal{
				Name: "test",
				Ingredients: []meals.MealIngredient{
					{
						Name:     "ingredient 1",
						Quantity: 1,
					},
				},
			},
			expectedErrors: []string{"Ingredients"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := service.CreateMeal(&testCase.meal)

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
	ingredientRepository := memory.IngredientRepository{}

	service := meals.NewService(&mealRepository, &ingredientRepository)

	mealToCreate := meals.Meal{
		Name:  "New meal",
		Notes: "Something exciting",
		Tags: []meals.Tag{
			{
				Id:   1,
				Name: "Quick",
			},
		},
		Ingredients: []meals.MealIngredient{
			{
				Id:       1,
				Name:     "Cheese",
				Quantity: 25,
				Unit:     "g",
				IsMain:   true,
			},
		},
	}

	createdMeal, err := service.CreateMeal(&mealToCreate)

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
	ingredientRepository := memory.IngredientRepository{}

	service := meals.NewService(&mealRepository, &ingredientRepository)

	type testCase struct {
		name           string
		meal           meals.Meal
		expectedErrors []string
	}

	testCases := []testCase{
		{
			name:           "Empty meal",
			meal:           meals.Meal{},
			expectedErrors: []string{"Name", "Ingredients"},
		},
		{
			name: "Ingredient with zero quantity",
			meal: meals.Meal{
				Name: "test",
				Ingredients: []meals.MealIngredient{
					{
						Id:       43,
						Name:     "ingredient 1",
						Quantity: 0,
					},
					{
						Name:     "ingredient 2",
						Quantity: 20,
						IsMain:   true,
					},
				},
			},
			expectedErrors: []string{"Ingredients.0"},
		},
		{
			name: "No main ingredient",
			meal: meals.Meal{
				Name: "test",
				Ingredients: []meals.MealIngredient{
					{
						Name:     "ingredient 1",
						Quantity: 1,
					},
				},
			},
			expectedErrors: []string{"Ingredients"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := service.UpdateMeal(&testCase.meal)

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
	ingredientRepository := memory.IngredientRepository{}

	service := meals.NewService(&mealRepository, &ingredientRepository)

	mealToUpdate := meals.Meal{
		Id:   23,
		Name: "New name",
		Tags: []meals.Tag{
			{
				Id:   1,
				Name: "Quick",
			},
		},
		Ingredients: []meals.MealIngredient{
			{
				Id:       1,
				Name:     "Cheese",
				Quantity: 25,
				Unit:     "g",
				IsMain:   true,
			},
		},
	}

	err := service.UpdateMeal(&mealToUpdate)

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

func TestServiceCanUpdateAnIngredient(t *testing.T) {
	mealRepository := memory.MealRepository{}
	ingredientRepository := memory.IngredientRepository{
		Store: []meals.Ingredient{
			{
				Id:   13,
				Name: "Tomates",
			},
		},
	}

	service := meals.NewService(&mealRepository, &ingredientRepository)

	ingredientToUpdate := meals.Ingredient{
		Id:   13,
		Name: "Tomatoes",
	}

	err := service.UpdateIngredient(&ingredientToUpdate)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(ingredientToUpdate.Errors) > 0 {
		t.Errorf("Unexpected validation errors: %v", ingredientToUpdate.Errors)
	}

	if len(ingredientRepository.Store) != 1 {
		t.Errorf("Expected 1 ingredient in the store, found %d", len(ingredientRepository.Store))
	}

	if ingredientRepository.Store[0].Name != ingredientToUpdate.Name {
		t.Errorf("Name not updated in store: expected %s, found %s", ingredientToUpdate.Name, ingredientRepository.Store[0].Name)
	}

}
func TestIngredientNameShouldBeUniquePerUser(t *testing.T) {
	mealRepository := memory.MealRepository{
		Store: []meals.Meal{
			{
				Id:     2,
				Name:   "A",
				UserId: 1,
				Ingredients: []meals.MealIngredient{
					{
						Id:       23,
						Name:     "Eggs",
						Quantity: 3,
						IsMain:   true,
					},
				},
			},
			{
				Id:     12,
				Name:   "B",
				UserId: 2,
				Ingredients: []meals.MealIngredient{
					{
						Id:       2,
						Name:     "Eggs",
						Quantity: 1,
						IsMain:   true,
					},
				},
			},
			{
				Id:     13,
				Name:   "C",
				UserId: 2,
				Ingredients: []meals.MealIngredient{
					{
						Id:       24,
						Name:     "Ham",
						Quantity: 1,
						IsMain:   true,
					},
				},
			},
		},
	}
	ingredientRepository := memory.IngredientRepository{
		Store: []meals.Ingredient{
			{
				Id:     2,
				UserId: 2,
				Name:   "Eggs",
			},
			{
				Id:     23,
				UserId: 1,
				Name:   "Eggs",
			},
			{
				Id:     24,
				UserId: 2,
				Name:   "Ham",
			},
		},
	}

	existingMeal := mealRepository.Store[0]
	mealToUpdate := mealRepository.Store[2]

	service := meals.NewService(&mealRepository, &ingredientRepository)

	newMeal := meals.Meal{
		Name:   "D",
		UserId: 1,
		Ingredients: []meals.MealIngredient{
			{
				Name:     "Eggs",
				Quantity: 1,
				IsMain:   true,
			},
		},
	}

	createdMeal, err := service.CreateMeal(&newMeal)

	if err != nil {
		t.Errorf("Unexpected error creating meal: %v", err)
	}

	if createdMeal.Ingredients[0].Id != existingMeal.Ingredients[0].Id {
		t.Errorf("Ingredient Ids should match but they don't (%d - %d)", newMeal.Ingredients[0].Id, existingMeal.Ingredients[0].Id)
	}

	updateMeal := meals.Meal{
		Id:     13,
		Name:   "C",
		UserId: 2,
		Ingredients: []meals.MealIngredient{
			{
				Name:     "Eggs",
				Quantity: 1,
				IsMain:   true,
			},
		},
	}

	err = service.UpdateMeal(&updateMeal)

	if err != nil {
		t.Errorf("Unexpected error creating meal: %v", err)
	}

	updatedMeal, _ := mealRepository.Get(13)

	if updatedMeal.Ingredients[0].Id != mealToUpdate.Ingredients[0].Id {
		t.Errorf("Ingredient Ids should match but they don't (%d - %d)", updatedMeal.Ingredients[0].Id, mealToUpdate.Ingredients[0].Id)
	}

}
