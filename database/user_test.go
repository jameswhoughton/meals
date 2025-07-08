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

func TestDatabaseUserRepository(t *testing.T) {
	init := func() (meals.UserRepository, func()) {
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
		return database.NewUserRespository(conn), closeDown
	}

	contract := contracts.UserRepository{
		Repo: init,
	}

	contract.Test(t)

}
