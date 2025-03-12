package meals

import (
	"fmt"
)

type Ingredient struct {
	Id     int
	UserId int
	Name   string
}

type IngredientFilter struct {
	Name   *string
	UserId int
}

type ErrorIngredientFilterInvalid struct{}

func (e ErrorIngredientFilterInvalid) Error() string {
	return "User ID must be set"
}

// Results should always be limited to a single user
func (f IngredientFilter) Validate() error {
	if f.UserId == 0 {
		return ErrorIngredientFilterInvalid{}
	}

	return nil
}

type ErrorIngredientNotFound struct {
	id int
}

func (e ErrorIngredientNotFound) Error() string {
	return fmt.Sprintf("Ingredient with the id: %d does not exist.", e.id)
}

type IngredientRepository interface {
	Get(id int) (Ingredient, error)
	List(filter IngredientFilter) ([]Ingredient, error)
	Create(ingredient Ingredient) (Ingredient, error)
	Update(ingredient Ingredient) error
	Destroy(id int) error
}
