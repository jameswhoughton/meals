package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
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

func (mr *MealRepository) getIngredientsForMeal(ctx context.Context, id int) ([]meals.MealIngredient, error) {
	query := `
	SELECT ingredient_id, i.name, quantity, unit, is_main
	FROM ingredients_meals im
	LEFT JOIN ingredients i
	ON im.ingredient_id = i.id
	WHERE im.meal_id = ?
	`

	var ingredients []meals.MealIngredient

	rows, err := mr.db.QueryContext(ctx, query, id)

	if err != nil {
		return ingredients, err
	}

	defer rows.Close()

	for rows.Next() {
		var ingredient meals.MealIngredient

		rows.Scan(&ingredient.Id, &ingredient.Name, &ingredient.Quantity, &ingredient.Unit, &ingredient.IsMain)

		ingredients = append(ingredients, ingredient)
	}

	return ingredients, nil
}

func (mr *MealRepository) getTagsForMeal(ctx context.Context, id int) ([]meals.Tag, error) {
	query := `
	SELECT tag_id, t.user_id, t.name
	FROM meals_tags mt
	LEFT JOIN tags t
	ON mt.tag_id = t.id
	WHERE mt.meal_id = ?
	`

	var tags []meals.Tag

	rows, err := mr.db.QueryContext(ctx, query, id)

	if err != nil {
		return tags, err
	}

	defer rows.Close()

	for rows.Next() {
		var tag meals.Tag

		rows.Scan(&tag.Id, &tag.UserId, &tag.Name)

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
		wheres = append(wheres, "AND name LIKE ?")
		values = append(values, "%"+*filter.Name+"%")
	}

	if len(filter.HasTags) > 0 {
		var params []string

		for _, v := range filter.HasTags {
			params = append(params, "?")
			values = append(values, v)
		}

		wheres = append(wheres, "AND tag_id IN ("+strings.Join(params, ", ")+")")

		query += `
		LEFT JOIN meals_tags t
		ON m.id = t.meal_id
		`
	}

	if len(filter.ExcludeMainIngredient) > 0 {
		var params []string

		for _, v := range filter.ExcludeMainIngredient {
			params = append(params, "?")
			values = append(values, v)
		}

		wheres = append(wheres, "AND (is_main = 1 AND ingredient_id NOT IN ("+strings.Join(params, ", ")+"))")

		query += `
		LEFT JOIN ingredients_meals i
		ON m.id = i.meal_id
		`
	}

	if filter.DateRange != nil {

		if filter.DateRange.Start != nil && filter.DateRange.End != nil {
			wheres = append(wheres, "AND (IFNULL(p.date, 0) > ? AND IFNULL(p.date, 0) < ?)")
			values = append(values, *filter.DateRange.Start, *filter.DateRange.End)
		} else if filter.DateRange.Start != nil {
			wheres = append(wheres, "AND IFNULL(p.date, 0) > ?")
			values = append(values, *filter.DateRange.Start)
		} else {
			wheres = append(wheres, "AND IFNULL(p.date, 0) < ?")
			values = append(values, *filter.DateRange.End)
		}
		query += `
		LEFT JOIN planner p
		ON m.id = p.meal_id
		`
	}

	query += "WHERE 1 = 1 " + strings.Join(wheres, " ")

	if len(filter.HasTags) > 0 {
		query += "GROUP BY m.id HAVING COUNT(DISTINCT tag_id) = " + strconv.Itoa(len(filter.HasTags))
	}

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

func createIngredient(ctx context.Context, tx *sql.Tx, userId int, name string) (int, error) {
	result, err := tx.ExecContext(ctx, "INSERT INTO ingredients (user_id, name) VALUES (?, ?)", userId, name)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func associateIngredientToMeal(ctx context.Context, tx *sql.Tx, mealId int, ingredient meals.MealIngredient) error {
	insertUpdateQuery := `
	REPLACE INTO ingredients_meals
	(ingredient_id, meal_id, quantity, unit, is_main)
	VALUES (?, ?, ?, ?, ?)
	`
	_, err := tx.ExecContext(
		ctx,
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

func removeOrphanedIngredients(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
	DELETE FROM ingredients
	WHERE id IN (
		SELECT i.id
		FROM ingredients i
		LEFT JOIN ingredients_meals im
		ON i.id = im.ingredient_id
		WHERE im.meal_id IS NULL
	)
	`)

	if err != nil {
		return err
	}

	return nil
}

func createTag(ctx context.Context, tx *sql.Tx, userId int, name string) (int, error) {
	result, err := tx.ExecContext(ctx, "INSERT INTO tags (user_id, name) VALUES (?, ?)", userId, name)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func associateTagToMeal(ctx context.Context, tx *sql.Tx, mealId int, tagId int) error {
	insertUpdateQuery := `
	REPLACE INTO meals_tags
	(tag_id, meal_id)
	VALUES (?, ?)
	`

	_, err := tx.ExecContext(
		ctx,
		insertUpdateQuery,
		tagId,
		mealId,
	)

	if err != nil {
		return err
	}

	return nil
}

func removeOrphanedTags(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
	DELETE FROM tags
	WHERE id IN (
		SELECT t.id
		FROM tags t
		LEFT JOIN meals_tags mt
		ON t.id = mt.tag_id
		WHERE mt.meal_id IS NULL
	)
	`)

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
		if ingredient.Id == 0 {
			id, err := createIngredient(ctx, tx, meal.UserId, ingredient.Name)

			if err != nil {
				return meal, fmt.Errorf("MealRepository.Create: Error inserting new ingredient: %v", err)
			}

			meal.Ingredients[i].Id = id
		}

		err = associateIngredientToMeal(ctx, tx, meal.Id, meal.Ingredients[i])

		if err != nil {
			return meal, fmt.Errorf("MealRepository.Create: Error associating ingredient: %v", err)
		}
	}

	for i, tag := range meal.Tags {
		if tag.Id == 0 {
			id, err := createTag(ctx, tx, meal.UserId, tag.Name)

			if err != nil {
				return meal, fmt.Errorf("MealRepository.Create: Error inserting new tag: %v", err)
			}

			meal.Tags[i].Id = id
		}

		err = associateTagToMeal(ctx, tx, meal.Id, meal.Tags[i].Id)

		if err != nil {
			return meal, fmt.Errorf("MealRepository.Create: Error associating tag: %v", err)
		}
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
			id, err := createIngredient(ctx, tx, meal.UserId, ingredient.Name)

			if err != nil {
				return fmt.Errorf("MealRepository.Update: Error inserting new ingredient: %v", err)
			}

			meal.Ingredients[i].Id = id
		}

		err = associateIngredientToMeal(ctx, tx, meal.Id, meal.Ingredients[i])

		if err != nil {
			return fmt.Errorf("MealRepository.Update: Error associating ingredient: %v", err)
		}

		deleteParams = append(deleteParams, meal.Ingredients[i].Id)
	}

	// Remove any ingredients no longer associated with the meal
	deleteIngredientsQuery := `
	DELETE FROM ingredients_meals
	WHERE meal_id = ?
	`

	if len(deleteParams) > 1 {
		deleteIngredientsQuery += `AND ingredient_id NOT IN (?` + strings.Repeat(",?", len(deleteParams)-2) + `)`
	}

	_, err = tx.ExecContext(ctx, deleteIngredientsQuery, deleteParams...)

	if err != nil {
		return fmt.Errorf("MealRepository.Update: Error disassociating ingredient: %v", err)
	}

	// Remove any ingredients that are no longer associated with any meal
	err = removeOrphanedIngredients(ctx, tx)

	deleteParams = []any{meal.Id}

	for i, tag := range meal.Tags {
		if tag.Id == 0 {
			id, err := createTag(ctx, tx, meal.UserId, tag.Name)

			if err != nil {
				return fmt.Errorf("MealRepository.Update: Error inserting new tag: %v", err)
			}

			meal.Tags[i].Id = id
		}

		err = associateTagToMeal(ctx, tx, meal.Id, meal.Tags[i].Id)

		if err != nil {
			return fmt.Errorf("MealRepository.Update: Error associating tag: %v", err)
		}

		deleteParams = append(deleteParams, meal.Tags[i].Id)
	}

	// Remove any tags no longer associated with the meal
	deleteTagQuery := `
	DELETE FROM meals_tags
	WHERE meal_id = ?
	`

	if len(deleteParams) > 1 {
		deleteTagQuery += `AND tag_id NOT IN (?` + strings.Repeat(",?", len(deleteParams)-2) + `)`

	}

	_, err = tx.ExecContext(ctx, deleteTagQuery, deleteParams...)

	if err != nil {
		return fmt.Errorf("MealRepository.Update: Error disassociating tag: %v", err)
	}

	// Remove any tags that are no longer associated with any meal
	err = removeOrphanedTags(ctx, tx)

	if err != nil {
		return fmt.Errorf("MealRepository.Update: Error removing orphaned tags: %v", err)
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

	if _, err := tx.ExecContext(ctx, "DELETE FROM ingredients_meals WHERE meal_id = ?", id); err != nil {
		return fmt.Errorf("MealRepository.Destroy: Error removing ingredients from meal (%d): %v", id, err)
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM meals_tags WHERE meal_id = ?", id); err != nil {
		return fmt.Errorf("MealRepository.Destroy: Error removing tags from meal (%d): %v", id, err)
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM meals WHERE id = ?", id); err != nil {
		return fmt.Errorf("MealRepository.Destroy: Error removing meal (%d): %v", id, err)
	}

	removeOrphanedIngredients(ctx, tx)

	removeOrphanedTags(ctx, tx)

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("MealRepository.Destroy: Error committing transaction: %v", err)
	}

	return nil
}

func (mr *MealRepository) AssignToDate(ctx context.Context, id int, date time.Time) error {
	query := `
	INSERT OR REPLACE INTO planner 
	(meal_id, date)
	SELECT ?, ?
	WHERE NOT EXISTS (SELECT * FROM planner WHERE meal_id = ? AND date = ?)
	`
	_, err := mr.db.ExecContext(ctx, query, id, date, id, date)

	if err != nil {
		return fmt.Errorf("MealRepository.AssignToDate: Error assigning meal %d to date %v: %v", id, date, err)
	}

	return nil
}
