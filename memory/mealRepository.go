package memory

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/jameswhoughton/meals/internal/meals"
)

type MealRepository struct {
	Store    []meals.Meal
	Calendar map[int][]time.Time
}

func (mr *MealRepository) Get(ctx context.Context, id int) (meals.Meal, error) {
	for _, meal := range mr.Store {
		if meal.Id == id {
			return meal, nil
		}
	}

	return meals.Meal{}, meals.ErrorMealNotFound{Id: id}
}

func (mr *MealRepository) Find(ctx context.Context, filter meals.MealFilter) ([]meals.Meal, error) {
	var meals []meals.Meal
	err := filter.Validate()

	if err != nil {
		return meals, err
	}

	for _, meal := range mr.Store {
		if meal.UserId != filter.UserId {
			continue
		}

		if filter.Name != nil && !strings.Contains(strings.ToLower(meal.Name), strings.ToLower(*filter.Name)) {
			continue
		}

		if len(filter.HasTags) > 0 {
			var found bool

			for _, tag := range meal.Tags {
				if slices.Contains(filter.HasTags, tag.Id) {
					found = true
				}
			}

			if !found {
				continue
			}
		}

		if filter.ExcludeMainIngredient != nil {
			skip := false
			for _, ingredient := range meal.Ingredients {
				if ingredient.IsMain && slices.Contains(filter.ExcludeMainIngredient, ingredient.Id) {
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
	count := make(map[string]bool, 0)

	for _, meal := range mr.Store {
		for _, ingredient := range meal.Ingredients {
			count[ingredient.Name] = true
		}
	}

	return len(count) + 1
}

func (mr *MealRepository) getNextTagId() int {
	count := make(map[string]bool, 0)

	for _, meal := range mr.Store {
		for _, tag := range meal.Tags {
			count[tag.Name] = true
		}
	}

	return len(count) + 1
}

func (mr *MealRepository) Create(ctx context.Context, meal meals.Meal) (meals.Meal, error) {
	meal.Id = len(mr.Store) + 1

	ingredientId := mr.getNextIngredientId()

	for i, ingredient := range meal.Ingredients {
		if ingredient.Id == 0 {
			ingredient.Id = ingredientId
			ingredientId++

			meal.Ingredients[i] = ingredient
		}
	}

	tagId := mr.getNextTagId()

	for i, tag := range meal.Tags {
		if tag.Id == 0 {
			tag.Id = tagId
			tagId++

			meal.Tags[i] = tag
		}
	}

	mr.Store = append(mr.Store, meal)

	return meal, nil
}

func (mr *MealRepository) Update(ctx context.Context, meal meals.Meal) error {
	for i, existingMeal := range mr.Store {
		if existingMeal.Id == meal.Id {
			ingredientId := mr.getNextIngredientId()

			// Check for any new ingredients and add Ids where required
			for i, ingredient := range meal.Ingredients {
				if ingredient.Id == 0 {
					ingredient.Id = ingredientId
					ingredientId++

					meal.Ingredients[i] = ingredient
				}
			}

			tagId := mr.getNextTagId()

			// Check for any new tags and add Ids where required
			for i, tag := range meal.Tags {
				if tag.Id == 0 {
					tag.Id = tagId
					tagId++

					meal.Tags[i] = tag
				}
			}

			mr.Store[i] = meal
		}
	}

	return nil
}

func (mr *MealRepository) Destroy(ctx context.Context, id int) error {
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

func (mr *MealRepository) AssignToDate(ctx context.Context, id int, date time.Time) error {
	if _, ok := mr.Calendar[id]; ok {
		mr.Calendar[id] = append(mr.Calendar[id], date)
	} else {
		mr.Calendar[id] = []time.Time{date}
	}

	return nil
}
