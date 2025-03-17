package memory

import "github.com/jameswhoughton/meals/internal/meals"

type MealRepository struct {
	Store []meals.Meal
}

func (mr *MealRepository) Get(id int) (meals.Meal, error) {
	for _, meal := range mr.Store {
		if meal.Id == id {
			return meal, nil
		}
	}

	return meals.Meal{}, meals.ErrorMealNotFound{Id: meal.Id}
}

func (mr *MealRepository) List(filter meals.MealFilter) ([]meals.Meal, error) {
	
}
	Create(Meal) (Meal, error)
	Update(Meal) error
	Destroy(id int) error
	GetForDateRange(userId int, startDate, endDate time.Time) ([]Meal, error)
	AssignToDate(id int, date time.Time) error
