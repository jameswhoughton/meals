package database_test

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/jameswhoughton/meals/database"
	"github.com/jameswhoughton/meals/internal/meals"
)

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

	contract := meals.IngredientRepositoryContract{
		Repo: init,
	}

	contract.Test(t)

}
