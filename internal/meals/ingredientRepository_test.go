package meals_test

import (
	"errors"
	"testing"

	"github.com/jameswhoughton/meals/internal/meals"
)

type IngredientRepositoryContractTest struct {
	repo func() (meals.IngredientRepository, func())
}

func (i IngredientRepositoryContractTest) Test(t *testing.T) {
	t.Run("Can create get update and delete an ingredient", func(t *testing.T) {
		repo, closeDown := i.repo()
		defer closeDown()

		newIngredient := meals.Ingredient{
			Name:   "Onions",
			UserId: 1,
		}

		createdIngredient, err := repo.Create(newIngredient)

		if err != nil {
			t.Errorf("Creating ingredient: unexpected error: %v", err)
		}

		if createdIngredient.Id == 0 {
			t.Error("Expected non zero ID")
		}

		if newIngredient.Name != createdIngredient.Name {
			t.Errorf("Expected Name %s, got %s", newIngredient.Name, createdIngredient.Name)
		}

		updatedIngredient := meals.Ingredient{
			Id:   createdIngredient.Id,
			Name: "Spring onions",
		}

		err = repo.Update(updatedIngredient)

		if err != nil {
			t.Errorf("Updating ingredient: unexpected error: %v", err)
		}

		fetchedIngredient, err := repo.Get(createdIngredient.Id)

		if err != nil {
			t.Errorf("Fetching ingredient: unexpected error: %v", err)
		}

		if createdIngredient.Id != fetchedIngredient.Id {
			t.Errorf("Expected ID %d, got %d", createdIngredient.Id, fetchedIngredient.Id)
		}

		if fetchedIngredient.Name != updatedIngredient.Name {
			t.Errorf("Expected Name %s, got %s", fetchedIngredient.Name, updatedIngredient.Name)
		}

		err = repo.Destroy(fetchedIngredient.Id)

		if err != nil {
			t.Errorf("Destroying ingredient: unexpected error: %v", err)
		}

		_, err = repo.Get(fetchedIngredient.Id)

		if err == nil {
			t.Error("Expected error, got none")
		}

		if !errors.Is(err, meals.ErrorIngredientNotFound{}) {
			t.Errorf("Expected error of type %T, got %T (%v)", meals.ErrorIngredientNotFound{}, err, err)
		}

	})

	t.Run("Can filter a list of ingredients by name", func(t *testing.T) {
		repo, closeDown := i.repo()
		defer closeDown()

		repo.Create(meals.Ingredient{Name: "Garlic"})
		repo.Create(meals.Ingredient{Name: "Onion"})
		repo.Create(meals.Ingredient{Name: "Spring onion"})
		searchString := "Onio"
		ingredients, err := repo.List(meals.IngredientFilter{Name: &searchString})

		if err != nil {
			t.Errorf("List ingredients: Unexpected error: %v", err)
		}

		if len(ingredients) != 2 {
			t.Errorf("Expected 2 results, got %d", len(ingredients))
		}

	})

	t.Run("Expected error when trying to filter with an empty user id", func(t *testing.T) {})
}
