package database

import (
	"database/sql"
	"fmt"

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

func (ir *IngredientRepository) GetById(id int) (meals.Ingredient, error) {
	var ingredient meals.Ingredient

	err := ir.db.QueryRow("SELECT id, user_id, name FROM ingredients WHERE id = ?", id).Scan(
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

func (ir *IngredientRepository) Find(search string, userId int) ([]meals.Ingredient, error) {
	ingredients := []meals.Ingredient{}

	rows, err := ir.db.Query(
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

func (ir *IngredientRepository) Update(ingredient meals.Ingredient) error {
	ir.db.Exec("UPDATE ingredients SET name = ? WHERE id = ? AND user_id = ?", ingredient.Name, ingredient.Id, ingredient.UserId)

	return nil
}
