package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

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

func (mr *MealRepository) getIngredientsForMeal(ctx context.Context, id int) ([]meals.Ingredient, error) {
	query := `
	SELECT id, name, quantity, unit
	FROM meal_ingredients
	WHERE meal_id = ?
	`

	var ingredients []meals.Ingredient

	rows, err := mr.db.QueryContext(ctx, query, id)

	if err != nil {
		return ingredients, err
	}

	defer rows.Close()

	for rows.Next() {
		var ingredient meals.Ingredient

		rows.Scan(&ingredient.Id, &ingredient.Name, &ingredient.Quantity, &ingredient.Unit)

		ingredients = append(ingredients, ingredient)
	}

	return ingredients, nil
}

func (mr *MealRepository) getTagsForMeal(ctx context.Context, id int) ([]meals.Tag, error) {
	query := `
	SELECT id, name
	FROM meal_tags mt
	WHERE meal_id = ?
	`

	var tags []meals.Tag

	rows, err := mr.db.QueryContext(ctx, query, id)

	if err != nil {
		return tags, err
	}

	defer rows.Close()

	for rows.Next() {
		var tag meals.Tag

		rows.Scan(&tag.Id, &tag.Name)

		tags = append(tags, tag)
	}

	return tags, nil
}

func (mr *MealRepository) Get(ctx context.Context, id int) (meals.Meal, error) {
	var meal meals.Meal

	mealQuery := `
	SELECT id, name, notes, user_id FROM meals WHERE id = ?
	`

	err := mr.db.QueryRowContext(ctx, mealQuery, id).Scan(
		&meal.Id,
		&meal.Name,
		&meal.Notes,
		&meal.UserId,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return meals.Meal{}, meals.ErrorMealNotFound{Id: id}
		}

		return meals.Meal{}, fmt.Errorf("MealRepository.Get: query error: %v", err)
	}

	ingredients, err := mr.getIngredientsForMeal(ctx, id)

	if err != nil {
		return meal, fmt.Errorf("MealRepository.Get: Error retrieving ingredients for meal %d: %v", meal.Id, err)
	}

	meal.Ingredients = ingredients

	tags, err := mr.getTagsForMeal(ctx, id)

	if err != nil {
		return meal, fmt.Errorf("MealRepository.Get: Error retrieving tags for meal %d: %v", meal.Id, err)
	}

	meal.Tags = tags

	return meal, nil
}
func (mr *MealRepository) Find(ctx context.Context, filter meals.MealFilter) ([]meals.Meal, error) {
	err := filter.Validate()

	if err != nil {
		return []meals.Meal{}, err
	}

	wheres := []string{
		"AND m.user_id = ?",
	}
	values := []any{
		filter.UserId,
	}
	var mealList []meals.Meal

	query := `
	SELECT DISTINCT m.id, m.name, m.user_id
	FROM meals m
	`

	if filter.Name != nil {
		wheres = append(wheres, "AND m.name LIKE ?")
		values = append(values, "%"+*filter.Name+"%")
	}

	if len(filter.Tags) > 0 {
		var params []string

		for _, v := range filter.Tags {
			params = append(params, "?")
			values = append(values, v)
		}

		wheres = append(wheres, "AND t.name IN ("+strings.Join(params, ", ")+")")

		query += `
		LEFT JOIN meal_tags t
		ON m.id = t.meal_id
		`
	}

	query += "WHERE 1 = 1 " + strings.Join(wheres, " ")

	if len(filter.Tags) > 0 {
		query += "GROUP BY m.id HAVING COUNT(DISTINCT t.name) = " + strconv.Itoa(len(filter.Tags))
	}

	fmt.Println(query, values)

	rows, err := mr.db.QueryContext(ctx, query, values...)

	if err != nil {
		return mealList, fmt.Errorf("MealRepository.List: query error: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var meal meals.Meal

		if err = rows.Scan(&meal.Id, &meal.Name, &meal.UserId); err != nil {
			return mealList, fmt.Errorf("MealRepository.List: row parse error: %v", err)
		}

		tags, _ := mr.getTagsForMeal(ctx, meal.Id)

		meal.Tags = tags

		mealList = append(mealList, meal)
	}

	return mealList, nil
}

func createIngredient(ctx context.Context, tx *sql.Tx, mealId int, ingredient meals.Ingredient) (int, error) {
	result, err := tx.ExecContext(ctx, `
		INSERT INTO meal_ingredients
		(meal_id, name, quantity, unit)
		VALUES (?, ?, ?, ?)
	`, mealId, ingredient.Name, ingredient.Quantity, ingredient.Unit)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func updateIngredient(ctx context.Context, tx *sql.Tx, ingredient meals.Ingredient) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE meal_ingredients
		SET name = ?,
		quantity = ?,
		unit = ?
		WHERE id = ?
	`, ingredient.Name, ingredient.Quantity, ingredient.Unit, ingredient.Id)

	if err != nil {
		return err
	}

	return nil
}

func createTag(ctx context.Context, tx *sql.Tx, mealId int, name string) (int, error) {
	result, err := tx.ExecContext(ctx, "INSERT INTO meal_tags (meal_id, name) VALUES (?, ?)", mealId, name)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func updateTag(ctx context.Context, tx *sql.Tx, tag meals.Tag) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE meal_tags
		SET name = ?
		WHERE id = ?
	`, tag.Name, tag.Id)

	if err != nil {
		return err
	}

	return nil
}

func (mr *MealRepository) Create(ctx context.Context, meal meals.Meal) (meals.Meal, error) {
	tx, err := mr.db.Begin()

	if err != nil {
		return meal, fmt.Errorf("MealRepository.Create: Error creating transaction: %v", err)
	}

	defer tx.Rollback()

	insertQuery := `
	INSERT INTO meals
	(user_id, name, notes)
	VALUES (?, ?, ?)
	`

	result, err := tx.ExecContext(
		ctx,
		insertQuery,
		meal.UserId,
		meal.Name,
		meal.Notes,
	)

	if err != nil {
		return meal, fmt.Errorf("MealRepository.Create: Error inserting meal: %v", err)
	}

	id, err := result.LastInsertId()

	if err != nil {
		return meal, err
	}

	meal.Id = int(id)

	for i, ingredient := range meal.Ingredients {
		id, err := createIngredient(ctx, tx, meal.Id, ingredient)

		if err != nil {
			return meal, fmt.Errorf("MealRepository.Create: Error inserting new ingredient: %v", err)
		}

		meal.Ingredients[i].Id = id
	}

	for i, tag := range meal.Tags {
		id, err := createTag(ctx, tx, meal.Id, tag.Name)

		if err != nil {
			return meal, fmt.Errorf("MealRepository.Create: Error inserting new tag: %v", err)
		}

		meal.Tags[i].Id = id
	}

	if err := tx.Commit(); err != nil {
		return meal, fmt.Errorf("MealRepository.Create: Error committing transaction: %v", err)
	}

	return meal, nil

}

func (mr *MealRepository) Update(ctx context.Context, meal meals.Meal) error {
	tx, err := mr.db.Begin()

	if err != nil {
		return fmt.Errorf("MealRepository.Update: Error creating transaction: %v", err)
	}

	defer tx.Rollback()

	updateQuery := `
	UPDATE meals
	SET name = ?,
	notes = ?
	WHERE id = ?
	`

	_, err = tx.ExecContext(
		ctx,
		updateQuery,
		meal.Name,
		meal.Notes,
		meal.Id,
	)

	if err != nil {
		return fmt.Errorf("MealRepository.Update: Error updating meals row: %v", err)
	}

	deleteParams := []any{meal.Id}

	for i, ingredient := range meal.Ingredients {
		if ingredient.Id == 0 {
			id, err := createIngredient(ctx, tx, meal.Id, ingredient)

			if err != nil {
				return fmt.Errorf("MealRepository.Update: Error inserting new ingredient: %v", err)
			}

			meal.Ingredients[i].Id = id

			deleteParams = append(deleteParams, id)

			continue
		}

		err := updateIngredient(ctx, tx, ingredient)

		if err != nil {
			return fmt.Errorf("MealRepository.Update: Error updating existing ingredient: %v", err)
		}

		deleteParams = append(deleteParams, ingredient.Id)
	}

	// Remove any ingredients no longer associated with the meal
	deleteIngredientsQuery := `
	DELETE FROM meal_ingredients
	WHERE meal_id = ?
	`

	if len(deleteParams) > 1 {
		deleteIngredientsQuery += `AND id NOT IN (?` + strings.Repeat(",?", len(deleteParams)-2) + `)`
	}

	_, err = tx.ExecContext(ctx, deleteIngredientsQuery, deleteParams...)

	if err != nil {
		return fmt.Errorf("MealRepository.Update: Error deleting ingredient: %v", err)
	}

	deleteParams = []any{meal.Id}

	for i, tag := range meal.Tags {
		if tag.Id == 0 {
			id, err := createTag(ctx, tx, meal.UserId, tag.Name)

			if err != nil {
				return fmt.Errorf("MealRepository.Update: Error inserting new tag: %v", err)
			}

			meal.Tags[i].Id = id

			deleteParams = append(deleteParams, id)
		}

		err := updateTag(ctx, tx, tag)

		if err != nil {
			return fmt.Errorf("MealRepository.Update: Error updating existing tag: %v", err)
		}

		deleteParams = append(deleteParams, tag.Id)
	}

	// Remove any tags no longer associated with the meal
	deleteTagQuery := `
	DELETE FROM meal_tags
	WHERE meal_id = ?
	`

	if len(deleteParams) > 1 {
		deleteTagQuery += `AND id NOT IN (?` + strings.Repeat(",?", len(deleteParams)-2) + `)`

	}

	_, err = tx.ExecContext(ctx, deleteTagQuery, deleteParams...)

	if err != nil {
		return fmt.Errorf("MealRepository.Update: Error deleting tag: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("MealRepository.Update: Error committing transaction: %v", err)
	}

	return nil
}

func (mr *MealRepository) Destroy(ctx context.Context, id int) error {
	tx, err := mr.db.Begin()

	if err != nil {
		return fmt.Errorf("MealRepository.Destroy: Error creating transaction: %v", err)
	}

	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "DELETE FROM planner WHERE meal_id = ?", id); err != nil {
		return fmt.Errorf("MealRepository.Destroy: Error removing meal (%d) from planner: %v", id, err)
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM meal_ingredients WHERE meal_id = ?", id); err != nil {
		return fmt.Errorf("MealRepository.Destroy: Error removing ingredients from meal (%d): %v", id, err)
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM meal_tags WHERE meal_id = ?", id); err != nil {
		return fmt.Errorf("MealRepository.Destroy: Error removing tags from meal (%d): %v", id, err)
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM meals WHERE id = ?", id); err != nil {
		return fmt.Errorf("MealRepository.Destroy: Error removing meal (%d): %v", id, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("MealRepository.Destroy: Error committing transaction: %v", err)
	}

	return nil
}

func (mr *MealRepository) FindIngredientNames(ctx context.Context, searchString string) ([]string, error) {
	ingredients := make([]string, 0)

	rows, err := mr.db.QueryContext(
		ctx,
		`SELECT DISTINCT name FROM meal_ingredients WHERE name LIKE ?`,
		"%"+searchString+"%",
	)
	if err != nil {
		return ingredients, fmt.Errorf("MealRepository.FindIngredientNames: query error: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var name string

		if err = rows.Scan(&name); err != nil {
			return ingredients, fmt.Errorf("MealRepository.FindIngredientNames: row parse error: %v", err)
		}

		ingredients = append(ingredients, name)
	}

	return ingredients, nil
}

func (mr *MealRepository) FindTagNames(ctx context.Context, searchString string) ([]string, error) {
	tags := make([]string, 0)

	rows, err := mr.db.QueryContext(
		ctx,
		`SELECT DISTINCT name FROM meal_tags WHERE name LIKE ?`,
		"%"+searchString+"%",
	)
	if err != nil {
		return tags, fmt.Errorf("MealRepository.FindTagNames: query error: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var name string

		if err = rows.Scan(&name); err != nil {
			return tags, fmt.Errorf("MealRepository.FindTagNames: row parse error: %v", err)
		}

		tags = append(tags, name)
	}

	return tags, nil
}

func (mr *MealRepository) TagNamesForUser(ctx context.Context, userId int) ([]string, error) {
	tags := make([]string, 0)

	rows, err := mr.db.QueryContext(
		ctx,
		`SELECT DISTINCT mt.name FROM meal_tags mt LEFT JOIN meals m ON mt.meal_id = m.id WHERE m.user_id = ?`,
		userId,
	)
	if err != nil {
		return tags, fmt.Errorf("MealRepository.TagNamesForUser: query error: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var name string

		if err = rows.Scan(&name); err != nil {
			return tags, fmt.Errorf("MealRepository.TagNamesForUser: row parse error: %v", err)
		}

		tags = append(tags, name)
	}

	return tags, nil
}
