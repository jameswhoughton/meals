package meals

import (
	"fmt"
)

/*
Specification:
- Has a single owner
- Can belong to multiple meals
- Cannot exist if it isn't associated with a meal
- Name should be unique
*/
type Ingredient struct {
	Id     int               `json:"id"`
	UserId int               `json:"-"`
	Name   string            `json:"name"`
	Errors map[string]string `json:"-"`
}

func (m *Ingredient) Validate() bool {
	m.Errors = make(map[string]string)

	if m.Name == "" {
		m.Errors["Name"] = "Name cannot be blank"
	}

	return len(m.Errors) == 0
}

type ErrorIngredientNotFound struct {
	Id int
}

func (e ErrorIngredientNotFound) Error() string {
	return fmt.Sprintf("Ingredient with the id: %d does not exist.", e.Id)
}

type IngredientRepository interface {
	Find(search string, userId int) ([]Ingredient, error)
	GetById(id int) (Ingredient, error)
	Update(ingredient Ingredient) error
	FromNames(names []string, userId int) (map[string]int, error)
}
