package meals

import (
	"fmt"
	"time"
)

type MealAttributes struct {
	Quick  bool
	Family bool
	Easy   bool
}

type MealIngredient struct {
	Id           int
	IngredientId int
	Name         string
	Quantity     int
	Unit         string
	IsMain       bool
}

type Meal struct {
	Id          int
	UserId      int
	Name        string
	Notes       string
	Attributes  MealAttributes
	Ingredients []MealIngredient
	LastEatenOn time.Time
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
	ExcludeIngredient *[]int
	DateRange         *DateRange
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

type MealRepository interface {
	Get(id int) (Meal, error)
	List(filter MealFilter) ([]Meal, error)
	Create(meal Meal) (Meal, error)
	Update(meal Meal) error
	Destroy(id int) error
	AssignToDate(id int, date time.Time) error
}
