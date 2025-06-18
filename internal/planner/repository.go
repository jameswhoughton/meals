package planner

import (
	"context"
	"errors"
	"time"
)

var ErrMealNotSet = errors.New("No meal set on date")

type Meal struct {
	Id     int
	UserId int
	Name   string
}

type Ingredient struct {
	Name     string
	Quantity int
	Unit     string
}

type Repository interface {

	// Returns the meal set on the given date for the given userId.
	//
	// Returns ErrMealNotSet if there is no meal assigned to the given date.
	// The time portion of the date is ignored.
	Get(ctx context.Context, date time.Time, userId int) (Meal, error)

	// Assign a mealId to the given date
	//
	// The time portion of the date is ignored.
	Add(ctx context.Context, date time.Time, mealId int) error

	// Remove any meals assigned to the date by the userID
	//
	// The time portion of the date is ignored.
	Clear(ctx context.Context, date time.Time, userId int) error

	// Get the combined ingredients for all meals assigned by the userId
	// in the date range.
	//
	// Ingredients are grouped by name and unit.
	GetIngredients(ctx context.Context, startDate, endDate time.Time, userId int) ([]Ingredient, error)
}
