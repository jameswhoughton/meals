package database_test

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jameswhoughton/meals"
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

func seedMeal(conn *sql.DB, meal meals.Meal) {
	seedUser(conn, meal.UserId)

	result, err := conn.Exec(
		`INSERT INTO meals
		(id, user_id, name, notes)
		VALUES (?, ?, ?, ?)
	`, meal.Id, meal.UserId, meal.Name, meal.Notes)

	if err != nil {
		log.Fatalf("Unable to seed meal: %v", err)
	}

	mealId, _ := result.LastInsertId()

	for _, ingredient := range meal.Ingredients {
		_, err := conn.Exec(`
			INSERT INTO meal_ingredients
			(meal_id, name, quantity, unit)
			VALUES (?, ?, ?, ?)
		`, mealId, ingredient.Name, ingredient.Quantity, ingredient.Unit)

		if err != nil {
			log.Fatalf("Unable to seed meal meal=%+v: %v", meal, err)
		}
	}

	for _, tag := range meal.Tags {
		_, err := conn.Exec(`
			INSERT INTO meal_tags
			(meal_id, name)
			VALUES (?, ?)
		`, mealId, tag.Name)

		if err != nil {
			log.Fatalf("Unable to seed meal: %v", err)
		}
	}
}
