package database_test

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/jameswhoughton/meals/database"
	"github.com/jameswhoughton/meals/internal/meals"
)

func TestDatabasetagRepository(t *testing.T) {
	init := func(tags []meals.Tag) (meals.TagRepository, func()) {
		conn, err := sql.Open("sqlite3", "db")

		if err != nil {
			log.Fatal(err)
		}

		err = database.Migrate(conn)

		if err != nil {
			log.Fatal(err)
		}

		for _, tag := range tags {
			_, err := conn.Exec("INSERT INTO tags (id, user_id, name) VALUES (?, ?, ?)", tag.Id, tag.UserId, tag.Name)

			if err != nil {
				log.Fatalf("Error inserting test data: %v", err)
			}
		}

		closeDown := func() {
			os.Remove("db")
		}
		return database.NewTagRepository(conn), closeDown
	}

	contract := meals.TagRepositoryContract{
		Repo: init,
	}

	contract.Test(t)

}
