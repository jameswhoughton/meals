package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jameswhoughton/meals/internal/meals"
)

type MealRepository struct {
	db *sql.DB
}

func (mr *MealRepository) Get(id int) (meals.Meal, error) {
	var meal meals.Meal

	mealQuery := `
	SELECT id, name, notes, user_id, quick, family, easy FROM meals WHERE id = ?
	`

	err := mr.db.QueryRow(mealQuery, id).Scan(
		&meal.Id,
		&meal.Name,
		&meal.Notes,
		&meal.UserId,
		&meal.Attributes.Quick,
		&meal.Attributes.Family,
		&meal.Attributes.Easy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return meals.Meal{}, meals.ErrorMealNotFound{Id: id}
		}

		return meals.Meal{}, fmt.Errorf("MealRepository.Get: query error: %v", err)
	}

	ingredientQuery := `
	SELECT ingredient_id, i.name, quantity, unit, is_main
	FROM ingredients_meals im
	LEFT JOIN ingredients i
	ON im.ingredient_id = i.id
	WHERE im.meal_id = ?
	`

	rows, err := mr.db.Query(ingredientQuery, meal.Id)

	if err != nil {
		return meal, fmt.Errorf("MealRepository.Get: Error retrieving ingredients for meal %d", meal.Id)
	}

	defer rows.Close()

	for rows.Next() {
		var ingredient meals.MealIngredient

		rows.Scan(&ingredient.Id, &ingredient.Name, &ingredient.Quantity, &ingredient.Unit, &ingredient.IsMain)

		meal.Ingredients = append(meal.Ingredients, ingredient)
	}

	return meal, nil
}
func (mr *MealRepository) List(filter meals.MealFilter) ([]meals.Meal, error) {
	err := filter.Validate()

	if err != nil {
		return []meals.Meal{}, err
	}

	wheres := []string{
		"AND user_id = ?",
	}
	values := []any{
		filter.UserId,
	}
	var mealList []meals.Meal

	if filter.Quick != nil {
		wheres = append(wheres, "AND quick = ?")
		values = append(values, *filter.Quick)
	}

	if filter.Easy != nil {
		wheres = append(wheres, "AND easy = ?")
		values = append(values, *filter.Easy)
	}

	if filter.Family != nil {
		wheres = append(wheres, "AND family = ?")
		values = append(values, *filter.Family)
	}

	if len(filter.ExcludeIngredient) != 0 {
		var params []string

		for _, v := range filter.ExcludeIngredient {
			params = append(params, "?")
			values = append(values, v)
		}

		wheres = append(wheres, "AND meal_id NOT IN ("+strings.Join(params, ", ")+")")
	}

	query := "SELECT id, name, user_id FROM ingredients WHERE 1 = 1 " + strings.Join(wheres, " ")

	rows, err := mr.db.Query(query, values...)

	if err != nil {
		return mealList, fmt.Errorf("MealRepository.List: query error: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var meal meals.Meal

		if err = rows.Scan(&meal.Id, &meal.Name, &meal.UserId); err != nil {
			return mealList, fmt.Errorf("MealRepository.List: row parse error: %v", err)
		}

		mealList = append(mealList, meal)
	}

	return mealList, nil
}
func (mr *MealRepository) Create(meal meals.Meal) (meals.Meal, error) {}
func (mr *MealRepository) Update(meal meals.Meal) error               {}
func (mr *MealRepository) Destroy(id int) error {
	tx, err := mr.db.Begin()

	if err != nil {
		return fmt.Errorf("MealRepository.Destroy: Error creating transaction: %v", err)
	}

	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM schedule WHERE meal_id = ?", id); err != nil {
		return fmt.Errorf("MealRepository.Destroy: Error removing meal (%d) from schedule: %v", id, err)
	}

	if _, err := tx.Exec("DELETE FROM ingredients_meals WHERE meal_id = ?", id); err != nil {
		return fmt.Errorf("MealRepository.Destroy: Error removing ingredients from meal (%d): %v", id, err)
	}

	if _, err := tx.Exec("DELETE FROM meals WHERE id = ?", id); err != nil {
		return fmt.Errorf("MealRepository.Destroy: Error removing meal (%d): %v", id, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("MealRepository.Destroy: Error committing transaction: %v", err)
	}

	return nil
}
func (mr *MealRepository) AssignToDate(id int, date time.time) error {}
