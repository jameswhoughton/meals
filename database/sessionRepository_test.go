package database_test

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/jameswhoughton/meals/database"
	"github.com/jameswhoughton/meals/web"
)

func TestDatabaseSessionService(t *testing.T) {
	init := func() (web.SessionRepository, func()) {
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
		return database.NewSessionRepository(conn), closeDown
	}

	contract := web.SessionRepositoryContract{
		Repo: init,
	}

	contract.Test(t)

}
