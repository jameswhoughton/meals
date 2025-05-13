package database

import (
	"database/sql"
	"embed"
	"io/fs"

	"github.com/jameswhoughton/migrate"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func Migrate(conn *sql.DB) error {

	migrationFS := fs.FS(migrationFiles)

	migrationLog, err := migrate.NewLogMySQL(conn)

	if err != nil {
		return err
	}

	migrationsFiles, err := fs.Sub(migrationFS, "migrations")

	if err != nil {
		return err
	}

	err = migrate.Migrate(conn, migrationsFiles, &migrationLog)

	if err != nil {
		return err
	}

	return nil
}

func Rollback(conn *sql.DB) error {
	migrationFS := fs.FS(migrationFiles)

	migrationLog, err := migrate.NewLogMySQL(conn)

	if err != nil {
		return err
	}

	migrationsFiles, err := fs.Sub(migrationFS, "migrations")

	if err != nil {
		return err
	}

	err = migrate.Rollback(conn, migrationsFiles, &migrationLog)

	if err != nil {
		return err
	}

	return nil
}
