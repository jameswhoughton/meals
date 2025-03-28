package main

import (
	"database/sql"
	"log"

	"github.com/jameswhoughton/meals/database"
	"github.com/jameswhoughton/meals/internal/auth"
	"github.com/jameswhoughton/meals/internal/meals"
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
	mealRepository := database.NewMealRepository(conn)
	ingredientRepository := database.NewIngredientRepository(conn)

	userService := auth.NewUserService(userRepository, sessionRepsoitory, 3600)
	mealService := meals.NewService(mealRepository, ingredientRepository)

	server := web.NewServer(
		"8000",
		&userService,
		&mealService,
		mealRepository,
		ingredientRepository,
	)

	log.Fatal(server.Start())
}
