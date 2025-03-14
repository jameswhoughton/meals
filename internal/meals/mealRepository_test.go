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

type MealRepositoryContract struct {
	repo func() (meals.MealRepository, func())
}

func (i MealRepositoryContract) Test(t *testing.T) {
	t.Run("Can create get update and delete a meal", func(t *testing.T) {
		repo, closeDown := i.repo()
		defer closeDown()

		newMeal := meals.Meal{
			Name:   "Bolognese",
			UserId: 1,
			Attributes: meals.MealAttributes{
				Easy:   true,
				Quick:  false,
				Family: true,
			},
			Ingredients: []meals.MealIngredient{
				{
					Id:       1,
					Name:     "Beef mince",
					Quantity: 500,
					Unit:     "gram",
					IsMain:   true,
				},
				{
					Id:       2,
					Name:     "Tinned tomatoes",
					Quantity: 2,
					Unit:     "can",
					IsMain:   false,
				},
			},
		}

		createdMeal, err := repo.Create(newMeal)

		if err != nil {
			t.Errorf("Creating ingredient: unexpected error: %v", err)
		}

		if createdMeal.Id == 0 {
			t.Error("Expected non zero ID")
		}

		if newMeal.Name != createdMeal.Name {
			t.Errorf("Expected Name %s, got %s", newMeal.Name, createdMeal.Name)
		}

		updatedMeal := createdMeal

		updatedMeal.Attributes.Easy = false

		updatedMeal.Ingredients = append(updatedMeal.Ingredients, meals.MealIngredient{
			Id:       3,
			Name:     "Onion",
			Quantity: 2,
		})

		err = repo.Update(updatedMeal)

		if err != nil {
			t.Errorf("Updating meal: unexpected error: %v", err)
		}

		fetchedMeal, err := repo.Get(createdMeal.Id)

		if err != nil {
			t.Errorf("Fetching meal: unexpected error: %v", err)
		}

		if createdMeal.Id != fetchedMeal.Id {
			t.Errorf("Expected ID %d, got %d", createdMeal.Id, fetchedMeal.Id)
		}

		if fetchedMeal.Name != updatedMeal.Name {
			t.Errorf("Expected Name %s, got %s", fetchedMeal.Name, updatedMeal.Name)
		}

		err = repo.Destroy(fetchedMeal.Id)

		if err != nil {
			t.Errorf("Destroying meal: unexpected error: %v", err)
		}

		_, err = repo.Get(fetchedMeal.Id)

		if err == nil {
			t.Error("Expected error, got none")
		}

		if !errors.Is(err, meals.ErrorIngredientNotFound{Id: fetchedMeal.Id}) {
			t.Errorf("Expected error of type %T, got %T (%v)", meals.ErrorIngredientNotFound{}, err, err)
		}
	})

	t.Run("Can filter a list of meals", func(t *testing.T) {
		repo, closeDown := i.repo()
		defer closeDown()

	})

	t.Run("Must include UserId when filtering meals", func(t *testing.T) {
		repo, closeDown := i.repo()
		defer closeDown()

	})

	t.Run("filter field lastEatenBefore cannot be a future date", func(t *testing.T) {
		repo, closeDown := i.repo()
		defer closeDown()

	})

	t.Run("Can assign a meal to a given date", func(t *testing.T) {
		repo, closeDown := i.repo()
		defer closeDown()

	})

	t.Run("Can fetch meals for a given date range", func(t *testing.T) {
		repo, closeDown := i.repo()
		defer closeDown()

	})
}

func TestDatabaseMealRepository(t *testing.T) {
	init := func() (meals.IngredientRepository, func()) {
		conn, err := sql.Open("sqlite3", "meals.db")

		if err != nil {
			log.Fatal(err)
		}

		err = database.Migrate(conn)

		if err != nil {
			log.Fatal(err)
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

func TestMemoryMealRepository(t *testing.T) {
	init := func() (meals.IngredientRepository, func()) {
		return &memory.IngredientRepository{
			Store: []meals.Ingredient{},
		}, func() {}
	}

	contract := IngredientRepositoryContract{
		init,
	}

	contract.Test(t)

}
