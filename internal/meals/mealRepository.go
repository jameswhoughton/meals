package meals

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

type Ingredient struct {
	Id       int
	Name     string
	Quantity int
	Unit     string
}

/*
Specification:
- Has a single owner
- Has atleast one ingredient
- Has one main ingredient
- Can have one or more tags
- Meal names are not unique
- Associated ingredients must have a non-zero quantity
*/
type Meal struct {
	Id          int
	UserId      int
	Name        string
	Notes       string
	Tags        []Tag
	Ingredients []Ingredient
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Errors      map[string]string
}

func (m *Meal) Validate() bool {
	m.Errors = make(map[string]string)

	if m.Name == "" {
		m.Errors["Name"] = "Name cannot be blank"
	}

	for i, ingredient := range m.Ingredients {
		if ingredient.Quantity == 0 {
			m.Errors["Ingredients."+strconv.Itoa(i)] = "Ingredient quantity must be greater than zero"
		}
	}

	return len(m.Errors) == 0
}

type MealFilter struct {
	UserId int
	Name   *string
	Tags   []int
}

type ErrorMealFilterInvalid struct {
	message string
}

func (e ErrorMealFilterInvalid) Error() string {
	return e.message
}

func (mf MealFilter) Validate() error {
	if mf.UserId == 0 {
		return ErrorMealFilterInvalid{"UserID must be set"}
	}

	return nil
}

type ErrorMealNotFound struct {
	Id int
}

func (e ErrorMealNotFound) Error() string {
	return fmt.Sprintf("Meal with the id: %d does not exist.", e.Id)
}

type MealRepository interface {
	Get(ctx context.Context, id int) (Meal, error)
	Find(ctx context.Context, filter MealFilter) ([]Meal, error)
	Create(ctx context.Context, meal Meal) (Meal, error)
	Update(ctx context.Context, meal Meal) error
	Destroy(ctx context.Context, id int) error
	FindIngredientNames(ctx context.Context, searchString string) ([]string, error)
}
