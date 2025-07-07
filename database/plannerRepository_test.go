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

func TestDatabasePlannerRepository(t *testing.T) {
	init := func(testData []meals.Meal) (meals.PlannerRepository, func()) {
		conn, err := sql.Open("mysql", "root@tcp(127.0.0.1:8002)/meals?parseTime=true")

		if err != nil {
			log.Fatal(err)
		}

		err = database.Migrate(conn)

		if err != nil {
			log.Fatal(err)
		}

		for _, meal := range testData {
			seedMeal(conn, meal)
		}

		closeDown := func() {
			err := database.Rollback(conn)

			if err != nil {
				log.Fatal(err)
			}
		}
		return database.NewPlannerRepository(conn), closeDown
	}

	contract := contracts.PlannerRepository{
		Repo: init,
	}

	contract.Test(t)

}
