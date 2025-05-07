package database

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"

	"github.com/jameswhoughton/meals/internal/meals"
)

func NewIngredientRepository(db *sql.DB) *IngredientRepository {
	return &IngredientRepository{
		db: db,
	}
}

type IngredientRepository struct {
	db *sql.DB
}

func (ir *IngredientRepository) GetById(ctx context.Context, id int) (meals.Ingredient, error) {
	var ingredient meals.Ingredient

	err := ir.db.QueryRowContext(ctx, "SELECT id, user_id, name FROM ingredients WHERE id = ?", id).Scan(
		&ingredient.Id,
		&ingredient.UserId,
		&ingredient.Name,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return meals.Ingredient{}, meals.ErrorIngredientNotFound{Id: id}
		}

		return meals.Ingredient{}, fmt.Errorf("IngredientRepository.Get: query error: %v", err)
	}

	return ingredient, nil
}

func (ir *IngredientRepository) Find(ctx context.Context, search string, userId int) ([]meals.Ingredient, error) {
	ingredients := []meals.Ingredient{}

	rows, err := ir.db.QueryContext(
		ctx,
		"SELECT id, user_id, name FROM ingredients WHERE name LIKE ? AND user_id = ?",
		"%"+search+"%",
		userId,
	)

	if err != nil {
		return ingredients, fmt.Errorf("IngredientRepository.Find: query error: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var ingredient meals.Ingredient

		if err = rows.Scan(&ingredient.Id, &ingredient.UserId, &ingredient.Name); err != nil {
			return ingredients, fmt.Errorf("IngredientRepository.Find: row parse error: %v", err)
		}

		ingredients = append(ingredients, ingredient)
	}

	return ingredients, nil
}

func (ir *IngredientRepository) Update(ctx context.Context, ingredient meals.Ingredient) error {
	ir.db.Exec("UPDATE ingredients SET name = ? WHERE id = ? AND user_id = ?", ingredient.Name, ingredient.Id, ingredient.UserId)

	return nil
}

func (ir *IngredientRepository) FromNames(ctx context.Context, names []string, userId int) (map[string]int, error) {
	ingredientMap := make(map[string]int, len(names))
	var values []any

	for _, name := range names {
		values = append(values, name)
	}

	values = append(values, userId)
	params := slices.Repeat([]string{"?"}, len(names))

	rows, err := ir.db.QueryContext(
		ctx,
		"SELECT id, name FROM ingredients WHERE name IN ("+strings.Join(params, ", ")+") AND user_id = ?",
		values...,
	)

	if err != nil {
		return ingredientMap, fmt.Errorf("IngredientRepository.FromNames: query error: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var id int
		var name string

		if err = rows.Scan(&id, &name); err != nil {
			return ingredientMap, fmt.Errorf("IngredientRepository.FromNames: row parse error: %v", err)
		}

		ingredientMap[name] = id
	}

	return ingredientMap, nil
}
