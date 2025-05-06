package database_test

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/jameswhoughton/meals/database"
	"github.com/jameswhoughton/meals/internal/auth"
)

func TestDatabaseUserService(t *testing.T) {
	init := func() (auth.UserRepository, func()) {
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
		return database.NewUserRespository(conn), closeDown
	}

	contract := auth.UserRepositoryContract{
		Repo: init,
	}

	contract.Test(t)

}
