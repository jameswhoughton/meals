package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jameswhoughton/meals/internal/meals"
)

func NewMealRepository(db *sql.DB) *MealRepository {
	return &MealRepository{
		db: db,
	}
}

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

func createIngredient(tx *sql.Tx, userId int, name string) (int, error) {
	result, err := tx.Exec("INSERT INTO ingredients (user_id, name), VALUES (?, ?)", userId, name)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func associateIngredientToMeal(tx *sql.Tx, mealId int, ingredient meals.MealIngredient) error {
	insertUpdateQuery := `
	INSERT INTO ingredients_meals
	(ingredient_id, meal_id, quantity, unit, is_main)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT REPLACE
	`

	_, err := tx.Exec(
		insertUpdateQuery,
		ingredient.Id,
		mealId,
		ingredient.Quantity,
		ingredient.Unit,
		ingredient.IsMain,
	)

	if err != nil {
		return err
	}

	return nil
}

func removeOrphanedIngredients(tx *sql.Tx) error {
	_, err := tx.Exec(`
	DELETE FROM ingredients
	WHERE id IN (
		SELECT i.id
		FROM ingredients i
		LEFT JOIN ingredients_meals im
		WHERE im.meal_id IS NULL
	)
	`)

	if err != nil {
		return err
	}

	return nil
}

func (mr *MealRepository) Create(meal meals.Meal) (meals.Meal, error) {
	tx, err := mr.db.Begin()

	if err != nil {
		return meal, fmt.Errorf("MealRepository.Create: Error creating transaction: %v", err)
	}

	defer tx.Rollback()

	insertQuery := `
	INSERT INTO meals
	(user_id, name, notes, quick, family, easy)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := tx.Exec(
		insertQuery,
		meal.UserId,
		meal.Name,
		meal.Notes,
		meal.Attributes.Quick,
		meal.Attributes.Family,
		meal.Attributes.Easy,
	)

	if err != nil {
		return meal, fmt.Errorf("MealRepository.Create: Erorr inserting meal: %v", err)
	}

	id, err := result.LastInsertId()

	if err != nil {
		return meal, err
	}

	meal.Id = int(id)

	for i, ingredient := range meal.Ingredients {
		if ingredient.Id == 0 {
			id, err := createIngredient(tx, meal.UserId, ingredient.Name)

			if err != nil {
				return meal, fmt.Errorf("MealRepository.Create: Error inserting new ingredient: %v", err)
			}

			meal.Ingredients[i].Id = id
		}

		err = associateIngredientToMeal(tx, meal.Id, ingredient)

		if err != nil {
			return meal, fmt.Errorf("MealRepository.Create: Error associating ingredient: %v", err)
		}
	}

	return meal, nil

}

func (mr *MealRepository) Update(meal meals.Meal) error {
	tx, err := mr.db.Begin()

	if err != nil {
		return fmt.Errorf("MealRepository.Update: Error creating transaction: %v", err)
	}

	defer tx.Rollback()

	updateQuery := `
	UPDATE meals
	SET name = ?,
	notes = ?,
	quick = ?,
	family = ?,
	easy = ?,
	WHERE id = ?
	`

	_, err = tx.Exec(
		updateQuery,
		meal.Name,
		meal.Notes,
		meal.Attributes.Quick,
		meal.Attributes.Family,
		meal.Attributes.Easy,
		meal.Id,
	)

	if err != nil {
		return fmt.Errorf("MealRepository.Update: Error updating meals row: %v", err)
	}

	deleteParams := []any{meal.Id}

	for i, ingredient := range meal.Ingredients {
		if ingredient.Id == 0 {
			id, err := createIngredient(tx, meal.UserId, ingredient.Name)

			if err != nil {
				return fmt.Errorf("MealRepository.Update: Error inserting new ingredient: %v", err)
			}

			meal.Ingredients[i].Id = id
		}

		err = associateIngredientToMeal(tx, meal.Id, ingredient)

		if err != nil {
			return fmt.Errorf("MealRepository.Update: Error associating ingredient: %v", err)
		}

		deleteParams = append(deleteParams, ingredient.Id)
	}

	// Remove any ingredients no longer associated with the meal
	_, err = tx.Exec(`
	DELETE FROM ingredients_meals
	WHERE meal_id = ?
	AND ingrediend_id NOT IN (?`+strings.Repeat(",?", len(deleteParams)-2)+`)
	`, deleteParams...)

	if err != nil {
		return fmt.Errorf("MealRepository.Update: Error disassociating ingredient: %v", err)
	}

	// Remove any ingredients that are no longer associated with any meal
	err = removeOrphanedIngredients(tx)

	if err != nil {
		return fmt.Errorf("MealRepository.Update: Error removing orphaned ingredients: %v", err)
	}

	return nil
}

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

func (mr *MealRepository) AssignToDate(id int, date time.Time) error {
	return nil
}
