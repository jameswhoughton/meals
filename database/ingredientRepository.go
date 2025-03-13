package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jameswhoughton/meals/internal/meals"
)

func NewIngredientRepository(db *sql.DB) *IngredientRepository {
	return &IngredientRepository{db}
}

type IngredientRepository struct {
	db *sql.DB
}

func (i *IngredientRepository) Get(id int) (meals.Ingredient, error) {
	var ingredient meals.Ingredient

	query := "SELECT id, name, user_id FROM ingredients WHERE id = ?"

	err := i.db.QueryRow(query, id).Scan(&ingredient.Id, &ingredient.Name, &ingredient.UserId)

	if err != nil {
		if err == sql.ErrNoRows {
			return meals.Ingredient{}, meals.ErrorIngredientNotFound{Id: id}
		}

		return meals.Ingredient{}, fmt.Errorf("IngredientRepository.Get: query error: %v", err)
	}

	return ingredient, nil
}

func (i *IngredientRepository) List(filter meals.IngredientFilter) ([]meals.Ingredient, error) {
	err := filter.Validate()

	if err != nil {
		return []meals.Ingredient{}, err
	}

	wheres := []string{
		"AND user_id = ?",
	}
	values := []any{
		filter.UserId,
	}
	var ingredients []meals.Ingredient

	if filter.Name != nil {
		wheres = append(wheres, "AND name LIKE ?")
		values = append(values, "%"+*filter.Name+"%")
	}

	query := "SELECT id, name, user_id FROM ingredients WHERE 1 = 1 " + strings.Join(wheres, " ")

	rows, err := i.db.Query(query, values...)

	if err != nil {
		return ingredients, fmt.Errorf("IngredientRepository.List: query error: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var ingredient meals.Ingredient

		if err = rows.Scan(&ingredient.Id, &ingredient.Name, &ingredient.UserId); err != nil {
			return ingredients, fmt.Errorf("IngredientRepository.List: row parse error: %v", err)
		}

		ingredients = append(ingredients, ingredient)
	}

	return ingredients, nil
}

func (i *IngredientRepository) Create(ingredient meals.Ingredient) (meals.Ingredient, error) {
	result, err := i.db.Exec(`
		INSERT INTO ingredients 
		(name, user_id)
		VALUES (?, ?)
	`, ingredient.Name, ingredient.UserId)

	if err != nil {
		return meals.Ingredient{}, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return meals.Ingredient{}, err
	}

	ingredient.Id = int(id)

	return ingredient, nil
}

func (i *IngredientRepository) Update(ingredient meals.Ingredient) error {
	_, err := i.db.Exec(`
		UPDATE ingredients
		SET name = ?
		WHERE id = ?
	`, ingredient.Name, ingredient.Id)

	if err != nil {
		return fmt.Errorf("IngredientRepository.Update: update query error: %v", err)
	}

	return nil
}

func (i *IngredientRepository) Destroy(id int) error {
	_, err := i.db.Exec(`
		DELETE FROM ingredients WHERE id = ?
	`, id)

	if err != nil {
		return fmt.Errorf("IngredientRepository.Destroy: update query error: %v", err)
	}

	return nil
}
