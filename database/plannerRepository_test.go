package database_test

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/jameswhoughton/meals/database"
	"github.com/jameswhoughton/meals/internal/planner"
)

func TestDatabasePlannerRepository(t *testing.T) {
	init := func(meals []planner.Meal) (planner.Repository, func()) {
		conn, err := sql.Open("sqlite3", "meals.db")

		if err != nil {
			log.Fatal(err)
		}

		err = database.Migrate(conn)

		if err != nil {
			log.Fatal(err)
		}

		for _, meal := range meals {
			_, err := conn.Exec("INSERT INTO meals (id, user_id, name) VALUES (?, ?, ?)", meal.Id, meal.UserId, meal.Name)

			if err != nil {
				log.Fatalf("Error inserting test data: %v", err)
			}
		}

		closeDown := func() {
			os.Remove("meals.db")
		}
		return database.NewPlannerRepository(conn), closeDown
	}

	contract := planner.RepositoryContract{
		Repo: init,
	}

	contract.Test(t)

}
