package main

import (
	"database/sql"
	"log"

	"github.com/jameswhoughton/meals/database"
	"github.com/jameswhoughton/meals/internal/auth"
	"github.com/jameswhoughton/meals/web"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	conn, err := sql.Open("sqlite3", "meals.db")

	if err != nil {
		log.Fatal(err)
	}

	err = database.Migrate(conn)

	if err != nil {
		log.Fatal(err)
	}

	userRepository := database.NewUserRespository(conn)
	sessionRepsoitory := database.NewSessionRepository(conn)

	userService := auth.NewUserService(userRepository, sessionRepsoitory, 3600)

	server := web.NewServer("8000", &userService)

	log.Fatal(server.Start())
}
