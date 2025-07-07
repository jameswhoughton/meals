package database_test

import (
	"database/sql"
	"log"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jameswhoughton/meals"
	"github.com/jameswhoughton/meals/contracts"
	"github.com/jameswhoughton/meals/database"
)

func TestDatabaseMealRepository(t *testing.T) {
	init := func() (meals.MealRepository, func(userId int), func()) {
		conn, err := sql.Open("mysql", "root@tcp(127.0.0.1:8002)/meals")

		if err != nil {
			log.Fatal(err)
		}

		err = database.Migrate(conn)

		if err != nil {
			log.Fatal(err)
		}

		closeDown := func() {
			err := database.Rollback(conn)

			if err != nil {
				log.Fatal(err)
			}
		}
		return database.NewMealRepository(conn), func(userId int) { seedUser(conn, userId) }, closeDown
	}

	contract := contracts.MealRepository{
		Repo: init,
	}

	contract.Test(t)

}

func TestDatabaseMealMetaDataRepository(t *testing.T) {
	init := func(meals []meals.Meal) (meals.MealMetaDataRepository, func()) {
		conn, err := sql.Open("mysql", "root@tcp(127.0.0.1:8002)/meals")

		if err != nil {
			log.Fatal(err)
		}

		err = database.Migrate(conn)

		if err != nil {
			log.Fatal(err)
		}

		// Seed the data
		for _, meal := range meals {
			seedMeal(conn, meal)
		}

		closeDown := func() {
			err := database.Rollback(conn)

			if err != nil {
				log.Fatal(err)
			}
		}
		return database.NewMealMetaDataRepository(conn), closeDown
	}

	contract := contracts.MealMetaDataRepository{
		Repo: init,
	}

	contract.Test(t)

}
