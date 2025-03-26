package meals

import (
	"fmt"
	"strconv"
	"time"
)

type MealAttributes struct {
	Quick  bool
	Family bool
	Easy   bool
}

type Ingredient struct {
	Id     int
	UserId int
	Name   string
}

type MealIngredient struct {
	Id       int
	Name     string
	Quantity int
	Unit     string
	IsMain   bool
}

type Meal struct {
	Id          int
	UserId      int
	Name        string
	Notes       string
	Attributes  MealAttributes
	Ingredients []MealIngredient
	LastEatenOn time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Errors      map[string]string
}

func (m *Meal) Validate() bool {
	m.Errors = make(map[string]string)

	if m.Name == "" {
		m.Errors["Name"] = "Name cannot be blank"
	}

	if len(m.Ingredients) == 0 {
		m.Errors["Ingredients"] = "Meal must have at least one ingredient"
	}

	var mainIngredientCount int

	for i, ingredient := range m.Ingredients {
		if ingredient.IsMain {
			mainIngredientCount++
		}

		if ingredient.Quantity == 0 {
			m.Errors["Ingredients."+strconv.Itoa(i)] = "Ingredient quantity must be greater than zero"
		}

		if ingredient.Unit == "" {
			m.Errors["Ingredients."+strconv.Itoa(i)] = "Ingredient unit cannot be blank"
		}
	}

	if mainIngredientCount != 1 {
		m.Errors["Ingredients"] = "Meal must have one main ingredient"
	}

	return len(m.Errors) == 0
}

type DateRange struct {
	Start *time.Time
	End   *time.Time
}

type MealFilter struct {
	UserId int
	Quick  *bool
	Family *bool
	Easy   *bool
	// Only include meals eaten before the given date
	ExcludeMainIngredient []int
	DateRange             *DateRange
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

	if mf.DateRange != nil {
		if mf.DateRange.Start != nil && time.Now().Before(*mf.DateRange.Start) {
			return ErrorMealFilterInvalid{"DateRange.Start cannot be a future date"}
		}

		if mf.DateRange.End != nil && time.Now().Before(*mf.DateRange.End) {
			return ErrorMealFilterInvalid{"DateRange.End cannot be a future date"}
		}

		if mf.DateRange.Start != nil && mf.DateRange.End != nil && mf.DateRange.End.Before(*mf.DateRange.Start) {
			return ErrorMealFilterInvalid{"DateRange.End cannot become before DateRange.Start"}
		}
	}

	return nil
}

type ErrorMealNotFound struct {
	Id int
}

func (e ErrorMealNotFound) Error() string {
	return fmt.Sprintf("Meal with the id: %d does not exist.", e.Id)
}

type Repository interface {
	Get(id int) (Meal, error)
	List(filter MealFilter) ([]Meal, error)
	Create(meal Meal) (Meal, error)
	Update(meal Meal) error
	Destroy(id int) error
	AssignToDate(id int, date time.Time) error
	FindIngredients(search string, userId int) ([]Ingredient, error)
}
