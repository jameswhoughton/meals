package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/jameswhoughton/meals/database"
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

	userService := database.NewUserService(conn)
	sessionService := database.NewSessionService(conn)

	mux := http.NewServeMux()

	web.AddRoutes(mux, &userService, &sessionService)

	fmt.Println("listening on port :8000")

	http.ListenAndServe(":8000", mux)
}
