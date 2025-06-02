package database_test

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jameswhoughton/meals/internal/planner"
)

func seedUser(conn *sql.DB, userId int) {
	// If the user has already been seeded, return
	var existingUser int

	err := conn.QueryRow("SELECT id FROM users WHERE id = ?", userId).Scan(&existingUser)

	if existingUser == userId {
		return
	}

	email := fmt.Sprintf("user_%d@example.com", userId)

	_, err = conn.Exec("INSERT INTO users (id, name, email, password) VALUES (?, 'user', ?, 'password')", userId, email)

	if err != nil {
		log.Fatalf("Unable to seed user: %v", err)
	}
}

func seedPlannerMeal(conn *sql.DB, meal planner.Meal, ingredients []planner.Ingredient) {
	seedUser(conn, meal.UserId)

	_, err := conn.Exec(
		`INSERT INTO meals
		(id, user_id, name)
		VALUES (?, ?, ?)
	`, meal.Id, meal.UserId, meal.Name)

	if err != nil {
		log.Fatalf("Unable to seed meal: %v", err)
	}

	for _, ingredient := range ingredients {
		_, err := conn.Exec(`
			INSERT INTO meal_ingredients
			(meal_id, name, quantity, unit)
			VALUES (?, ?, ?, ?)
		`, meal.Id, ingredient.Name, ingredient.Quantity, ingredient.Unit)

		if err != nil {
			log.Fatalf("Unable to seed meal: %v", err)
		}
	}
}
