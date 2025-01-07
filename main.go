package main

import (
	"database/sql"
	"os"

	"github.com/jameswhoughton/migrate"
)

func setup() *sql.DB {
	// Directory containing migrations
	migrationDir := "migrations"

	// Create the connection to the DB
	db, _ := sql.Open("sqlite3", "meals.db")

	// Create an instance of the migration log
	log, err := migrate.NewLogSQLite(db)

	if err != nil {
		panic(err)
	}

	// Call Migrate to run migrations
	migrate.Migrate(db, os.DirFS(migrationDir), &log)

	return db
}

func main() {
	_ = setup()
}
