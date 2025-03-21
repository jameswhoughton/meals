package memory

import (
	"slices"
	"time"

	"github.com/jameswhoughton/meals/internal/meals"
)

type MealRepository struct {
	Store    []meals.Meal
	Calendar map[int][]time.Time
}

func (mr *MealRepository) Get(id int) (meals.Meal, error) {
	for _, meal := range mr.Store {
		if meal.Id == id {
			return meal, nil
		}
	}

	return meals.Meal{}, meals.ErrorMealNotFound{Id: id}
}

func (mr *MealRepository) List(filter meals.MealFilter) ([]meals.Meal, error) {
	var meals []meals.Meal
	err := filter.Validate()

	if err != nil {
		return meals, err
	}

	for _, meal := range mr.Store {
		if meal.UserId != filter.UserId {
			continue
		}

		if filter.Easy != nil && *filter.Easy != meal.Attributes.Easy {
			continue
		}

		if filter.Family != nil && *filter.Family != meal.Attributes.Family {
			continue
		}

		if filter.Quick != nil && *filter.Quick != meal.Attributes.Quick {
			continue
		}

		if filter.ExcludeIngredient != nil {
			skip := false
			for _, ingredient := range meal.Ingredients {
				if ingredient.IsMain && slices.Contains(filter.ExcludeIngredient, ingredient.Id) {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}

		if filter.DateRange != nil {
			if dates, ok := mr.Calendar[meal.Id]; ok {
				skip := false
				for _, date := range dates {
					if filter.DateRange.Start != nil && date.Before(*filter.DateRange.Start) {
						skip = true
						break
					}

					if filter.DateRange.End != nil && date.After(*filter.DateRange.End) {
						skip = true
						break
					}

				}

				if skip {
					continue
				}
			}
		}

		meals = append(meals, meal)
	}
	return meals, nil
}

// Loop over all ingredients across all meals to compute the next ID
func (mr *MealRepository) getNextIngredientId() int {
	var id int

	for _, meal := range mr.Store {
		id += len(meal.Ingredients)
	}

	return id + 1
}

func (mr *MealRepository) Create(meal meals.Meal) (meals.Meal, error) {
	meal.Id = len(mr.Store) + 1

	for i, ingredient := range meal.Ingredients {
		if ingredient.Id == 0 {
			ingredient.Id = mr.getNextIngredientId()

			meal.Ingredients[i] = ingredient
		}
	}

	mr.Store = append(mr.Store, meal)

	return meal, nil
}

func (mr *MealRepository) Update(meal meals.Meal) error {
	for i, existingMeal := range mr.Store {
		if existingMeal.Id == meal.Id {
			mr.Store[i] = meal
		}
	}

	return nil
}

func (mr *MealRepository) Destroy(id int) error {
	var meals []meals.Meal

	for _, meal := range mr.Store {
		if meal.Id == id {
			delete(mr.Calendar, id)

			continue
		}

		meals = append(meals, meal)
	}

	mr.Store = meals

	return nil
}

func (mr *MealRepository) AssignToDate(id int, date time.Time) error {
	if _, ok := mr.Calendar[id]; ok {
		mr.Calendar[id] = append(mr.Calendar[id], date)
	} else {
		mr.Calendar[id] = []time.Time{date}
	}

	return nil
}
