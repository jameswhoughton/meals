package meals

import (
	"context"
	"errors"
	"strconv"
	"time"
)

type Ingredient struct {
	Id       int
	Name     string
	Quantity int
	Unit     string
}

type Tag struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

var ErrMealNotFound = errors.New("meal not found")

/*
Specification:
- Has a single owner
- Has at least one ingredient
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
	Tags   []string
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

type Repository interface {

	// Return a meal matching the given ID
	//
	// If a meal ID does not exist, returns ErrMealNotFound
	Get(ctx context.Context, id int) (Meal, error)

	// Filter meals for a given user
	//
	// Meals can be filtered by name, tags and owner (user id)
	// The UserId filter is required,
	Find(ctx context.Context, filter MealFilter) ([]Meal, error)

	// Stores a new meal
	Create(ctx context.Context, meal Meal) (Meal, error)

	// Updates an existing meal
	//
	// Returns ErrMealNotFound id the meal.Id does not exist.
	Update(ctx context.Context, meal Meal) error

	// Deletes a meal
	Destroy(ctx context.Context, id int) error

	// Find ingredient names that partially match the searchString
	//
	// All ingredients are searched (regardless of owner).
	// Matches any part of the ingredient name e.g. searchString of `at` matches `tomato`.
	FindIngredientNames(ctx context.Context, searchString string) ([]string, error)

	// Find tag names that partially match the searchString
	//
	// All tags are searched (regardless of owner).
	// Matches any part of the tag name e.g. searchString of `ui` matches `quick`.
	FindTagNames(ctx context.Context, searchString string) ([]string, error)

	// Find units that partially match the searchString
	//
	// All units are searched (regardless of owner).
	// Matches are case-insensitive and match any part of the unit
	// e.g. searchString of `g` matches `KG`.
	FindUnitNames(ctx context.Context, searchString string) ([]string, error)

	// Returns all the tag names used by the given userId
	TagNamesForUser(ctx context.Context, userId int) ([]string, error)
}
