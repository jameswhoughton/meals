package meals

import (
	"context"
	"errors"
	"testing"
)

type IngredientRepositoryContract struct {
	// As the IngredientRepository is only responsible for fetching/editing existing ingredients,
	// any ingredients required for the test should be added directly to the store
	Repo func(ingredients []Ingredient) (IngredientRepository, func())
}

func (i IngredientRepositoryContract) Test(t *testing.T) {
	t.Run("Can filter a list of ingredients by name", func(t *testing.T) {
		ingredients := []Ingredient{
			{
				Id:     1,
				UserId: 1,
				Name:   "Onion",
			},
			{
				Id:     2,
				UserId: 1,
				Name:   "Spring onion",
			},
			{
				Id:     3,
				UserId: 2,
				Name:   "Red Onion",
			},
		}

		repo, closeDown := i.Repo(ingredients)
		defer closeDown()

		ctx := context.Background()

		searchString := "Onio"
		ingredients, err := repo.Find(ctx, searchString, 1)

		if err != nil {
			t.Errorf("List ingredients: Unexpected error: %v", err)
		}

		if len(ingredients) != 2 {
			t.Errorf("Expected 2 results, got %d", len(ingredients))
		}

	})

	t.Run("Can update the name of an ingredient", func(t *testing.T) {
		ingredients := []Ingredient{
			{
				Id:   1,
				Name: "Spring onin",
			},
		}

		repo, closeDown := i.Repo(ingredients)
		defer closeDown()

		ctx := context.Background()

		newName := "Spring onion"

		err := repo.Update(ctx, Ingredient{Id: 1, Name: newName})

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		fetchedIngredient, err := repo.GetById(ctx, 1)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if fetchedIngredient.Name != newName {
			t.Errorf("Expected ingredient name to be updated to %s, found %s", newName, fetchedIngredient.Name)
		}
	})

	t.Run("Can Fetch ingredient by ID", func(t *testing.T) {
		ingredients := []Ingredient{
			{
				Id:   1,
				Name: "Apple",
			},
			{
				Id:   2,
				Name: "Cheese",
			},
		}

		repo, closeDown := i.Repo(ingredients)
		defer closeDown()

		ctx := context.Background()

		ingredient, err := repo.GetById(ctx, 1)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if ingredient.Name != "Apple" {
			t.Errorf("Expected ingredient with name 'apple' got '%s'", ingredient.Name)
		}

		_, err = repo.GetById(ctx, 10)

		if err == nil {
			t.Errorf("Expected error, got none")
		}

		if !errors.As(err, &ErrorIngredientNotFound{Id: 10}) {
			t.Errorf("Expected error of type %T, got %T (%v)", ErrorIngredientNotFound{}, err, err)
		}

	})

	t.Run("Can fetch ingredient IDs by name", func(t *testing.T) {

		ingredients := []Ingredient{
			{
				Id:     1,
				UserId: 1,
				Name:   "Apple",
			},
			{
				Id:     2,
				UserId: 1,
				Name:   "Cheese",
			},
			{
				Id:     3,
				UserId: 2,
				Name:   "Ham",
			},
		}

		repo, closeDown := i.Repo(ingredients)
		defer closeDown()

		ctx := context.Background()

		ingredientMap, err := repo.FromNames(ctx, []string{"Apple", "Cheese", "Ham"}, 1)

		if err != nil {
			t.Errorf("Unexpected error fetching ingredient map: %v", err)
		}

		if ingredientMap["Apple"] != 1 {
			t.Errorf("Incorrect Id for 'Apple', should be 1, found %d", ingredientMap["Apple"])
		}

		if ingredientMap["Cheese"] != 2 {
			t.Errorf("Incorrect Id for 'Cheese', should be 2, found %d", ingredientMap["Cheese"])
		}

		if ingredientMap["Ham"] != 0 {
			t.Errorf("Incorrect Id for 'Ham', should be 0, found %d", ingredientMap["Ham"])
		}
	})
}
