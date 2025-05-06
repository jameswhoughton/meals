package database_test

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/jameswhoughton/meals/database"
	"github.com/jameswhoughton/meals/internal/meals"
)

func TestDatabaseRepository(t *testing.T) {
	init := func() (meals.MealRepository, func()) {
		conn, err := sql.Open("sqlite3", "db")

		if err != nil {
			log.Fatal(err)
		}

		err = database.Migrate(conn)

		if err != nil {
			log.Fatal(err)
		}

		closeDown := func() {
			os.Remove("db")
		}
		return database.NewMealRepository(conn), closeDown
	}

	contract := meals.MealRepositoryContract{
		Repo: init,
	}

	contract.Test(t)

}
