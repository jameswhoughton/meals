package meals_test

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"testing"

	"github.com/jameswhoughton/meals/database"
	"github.com/jameswhoughton/meals/internal/meals"
	"github.com/jameswhoughton/meals/memory"
)

type IngredientRepositoryContract struct {
	// As the IngredientRepository is only responsible for fetching/editing existing ingredients,
	// any ingredients required for the test should be added directly to the store
	repo func(ingredients []meals.Ingredient) (meals.IngredientRepository, func())
}

func (i IngredientRepositoryContract) Test(t *testing.T) {
	t.Run("Can filter a list of ingredients by name", func(t *testing.T) {
		ingredients := []meals.Ingredient{
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

		repo, closeDown := i.repo(ingredients)
		defer closeDown()

		searchString := "Onio"
		ingredients, err := repo.Find(searchString, 1)

		if err != nil {
			t.Errorf("List ingredients: Unexpected error: %v", err)
		}

		if len(ingredients) != 2 {
			t.Errorf("Expected 2 results, got %d", len(ingredients))
		}

	})

	t.Run("Can update the name of an ingredient", func(t *testing.T) {
		ingredients := []meals.Ingredient{
			{
				Id:   1,
				Name: "Spring onin",
			},
		}

		repo, closeDown := i.repo(ingredients)
		defer closeDown()

		newName := "Spring onion"

		err := repo.Update(meals.Ingredient{Id: 1, Name: newName})

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		fetchedIngredient, err := repo.GetById(1)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if fetchedIngredient.Name != newName {
			t.Errorf("Expected ingredient name to be updated to %s, found %s", newName, fetchedIngredient.Name)
		}
	})

	t.Run("Can Fetch ingredient by ID", func(t *testing.T) {
		ingredients := []meals.Ingredient{
			{
				Id:   1,
				Name: "Apple",
			},
			{
				Id:   2,
				Name: "Cheese",
			},
		}

		repo, closeDown := i.repo(ingredients)
		defer closeDown()

		ingredient, err := repo.GetById(1)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if ingredient.Name != "Apple" {
			t.Errorf("Expected ingredient with name 'apple' got '%s'", ingredient.Name)
		}

		_, err = repo.GetById(10)

		if err == nil {
			t.Errorf("Expected error, got none")
		}

		if !errors.As(err, &meals.ErrorIngredientNotFound{Id: 10}) {
			t.Errorf("Expected error of type %T, got %T (%v)", meals.ErrorIngredientNotFound{}, err, err)
		}

	})
}

func TestDatabaseIngredientRepository(t *testing.T) {
	init := func(ingredients []meals.Ingredient) (meals.IngredientRepository, func()) {
		conn, err := sql.Open("sqlite3", "meals.db")

		if err != nil {
			log.Fatal(err)
		}

		err = database.Migrate(conn)

		if err != nil {
			log.Fatal(err)
		}

		for _, ingredient := range ingredients {
			_, err := conn.Exec("INSERT INTO ingredients (id, user_id, name) VALUES (?, ?, ?)", ingredient.Id, ingredient.UserId, ingredient.Name)

			if err != nil {
				log.Fatalf("Error inserting test data: %v", err)
			}
		}

		closeDown := func() {
			os.Remove("meals.db")
		}
		return database.NewIngredientRepository(conn), closeDown
	}

	contract := IngredientRepositoryContract{
		init,
	}

	contract.Test(t)

}

func TestMemoryIngredientRepository(t *testing.T) {
	init := func(ingredients []meals.Ingredient) (meals.IngredientRepository, func()) {
		return &memory.IngredientRepository{
			Store: ingredients,
		}, func() {}
	}

	contract := IngredientRepositoryContract{
		init,
	}

	contract.Test(t)

}
