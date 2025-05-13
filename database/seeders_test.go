package database_test

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jameswhoughton/meals/internal/planner"
)

func seedUser(conn *sql.DB, userId int) {
	email := fmt.Sprintf("user_%d@example.com", userId)

	_, err := conn.Exec("INSERT INTO users (id, name, email, password) VALUES (?, 'user', ?, 'password')", userId, email)

	if err != nil {
		log.Fatalf("Unable to seed user: %v", err)
	}
}

func seedPlannerMeal(conn *sql.DB, meal planner.Meal) {
	seedUser(conn, meal.UserId)

	_, err := conn.Exec(
		`INSERT INTO meals
		(id, user_id, name)
		VALUES (?, ?, ?)
	`, meal.Id, meal.UserId, meal.Name)

	if err != nil {
		log.Fatalf("Unable to seed meal: %v", err)
	}
}
