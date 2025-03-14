package meals

import "time"

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

type MealFilter struct {
	UserId int
	Quick  *bool
	Family *bool
	Easy   *bool
	// Only include meals eaten before the given date
	EatenBeforeDate *time.Time
	// Exclude any meals that have a main ingredient found in the given slice
	ExcludeIngredient *[]int
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

	if mf.EatenBeforeDate != nil && time.Now().Before(*mf.EatenBeforeDate) {
		return ErrorMealFilterInvalid{"EatenBeforeDate cannot be a future date"}
	}

	return nil
}

type MealRepository interface {
	Get(id int) (Meal, error)
	List(filter MealFilter) ([]Meal, error)
	Create(Meal) (Meal, error)
	Update(Meal) error
	Destroy(id int) error
	GetForDateRange(userId int, startDate, endDate time.Time) ([]Meal, error)
	AssignToDate(id int, date time.Time) error
}
