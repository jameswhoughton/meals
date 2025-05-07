package database_test

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/jameswhoughton/meals/database"
	"github.com/jameswhoughton/meals/internal/account"
)

func TestDatabaseAccountRepository(t *testing.T) {
	init := func() (account.Repository, func()) {
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
		return database.NewAccountRespository(conn), closeDown
	}

	contract := account.RepositoryContract{
		Repo: init,
	}

	contract.Test(t)

}
