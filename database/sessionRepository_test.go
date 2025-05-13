package database_test

import (
	"database/sql"
	"log"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jameswhoughton/meals/database"
	"github.com/jameswhoughton/meals/web"
)

func TestDatabaseSessionService(t *testing.T) {
	init := func() (web.SessionRepository, func()) {
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
		return database.NewSessionRepository(conn), closeDown
	}

	contract := web.SessionRepositoryContract{
		Repo: init,
	}

	contract.Test(t)

}
