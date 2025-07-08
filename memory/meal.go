package memory

import (
	"context"
	"maps"
	"slices"
	"strings"

	"github.com/jameswhoughton/meals"
)

type MealRepository struct {
	Store []meals.Meal
}

func (mr *MealRepository) Get(ctx context.Context, id int) (meals.Meal, error) {
	for _, meal := range mr.Store {
		if meal.Id == id {
			return meal, nil
		}
	}

	return meals.Meal{}, meals.ErrMealNotFound
}

func (mr *MealRepository) Find(ctx context.Context, filter meals.MealFilter) ([]meals.Meal, error) {
	var meals []meals.Meal

	for _, meal := range mr.Store {
		if meal.UserId != filter.UserId {
			continue
		}

		if filter.Name != nil && !strings.Contains(strings.ToLower(meal.Name), strings.ToLower(*filter.Name)) {
			continue
		}

		if len(filter.Tags) > 0 {
			var foundCount int

			for _, tag := range meal.Tags {
				if slices.Contains(filter.Tags, tag.Name) {
					foundCount++
				}
			}

			if len(filter.Tags) != foundCount {
				continue
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
			continue
		}

		meals = append(meals, meal)
	}

	mr.Store = meals

	return nil
}

type MealMetaDataRepository struct {
	Store []meals.Meal
}

func (mr *MealMetaDataRepository) FindIngredientNames(ctx context.Context, searchString string) ([]string, error) {
	names := make([]string, 0)

	for _, meal := range mr.Store {
		for _, ingredient := range meal.Ingredients {
			if !strings.Contains(strings.ToLower(ingredient.Name), strings.ToLower(searchString)) || slices.Contains(names, ingredient.Name) {
				continue
			}

			names = append(names, ingredient.Name)
		}
	}

	return names, nil
}

func (mr *MealMetaDataRepository) FindTagNames(ctx context.Context, searchString string) ([]string, error) {
	names := make([]string, 0)

	for _, meal := range mr.Store {
		for _, tag := range meal.Tags {
			if !strings.Contains(strings.ToLower(tag.Name), strings.ToLower(searchString)) || slices.Contains(names, tag.Name) {
				continue
			}

			names = append(names, tag.Name)
		}
	}

	return names, nil
}

func (mr *MealMetaDataRepository) FindUnitNames(ctx context.Context, searchString string) ([]string, error) {
	names := make([]string, 0)

	for _, meal := range mr.Store {
		for _, ingredient := range meal.Ingredients {
			if !strings.Contains(strings.ToLower(ingredient.Unit), strings.ToLower(searchString)) || slices.Contains(names, ingredient.Unit) {
				continue
			}

			names = append(names, ingredient.Unit)
		}
	}

	return names, nil
}

func (mr *MealMetaDataRepository) TagNamesForUser(ctx context.Context, userId int) ([]string, error) {
	tags := make([]string, 0)

	for _, meal := range mr.Store {
		if meal.UserId != userId {
			continue
		}

		for _, tag := range meal.Tags {
			if slices.Contains(tags, tag.Name) {
				continue
			}

			tags = append(tags, tag.Name)
		}
	}

	return tags, nil
}

func (mr *MealMetaDataRepository) findMeal(id int) (meals.Meal, error) {
	var meal meals.Meal

	for _, m := range mr.Store {
		if m.Id == id {
			meal = m
			break
		}
	}

	if meal.Id == 0 {
		return meal, meals.ErrMealNotFound
	}

	return meal, nil
}

func (mr *MealMetaDataRepository) GetTotalIngredients(ct context.Context, mealIds []int) ([]meals.IngredientTotal, error) {
	totals := make(map[string]meals.IngredientTotal)

	for _, mealId := range mealIds {
		if mealId == 0 {
			continue
		}

		meal, err := mr.findMeal(mealId)

		if err != nil {
			return nil, err
		}

		ingredients := meal.Ingredients

		for _, ingredient := range ingredients {
			totalsKey := ingredient.Name + "|" + ingredient.Unit

			total, ok := totals[totalsKey]

			if !ok {
				totals[totalsKey] = meals.IngredientTotal{
					Name:     ingredient.Name,
					Quantity: ingredient.Quantity,
					Unit:     ingredient.Unit,
				}

				continue
			}

			total.Quantity += ingredient.Quantity

			totals[totalsKey] = total
		}
	}

	return slices.Collect(maps.Values(totals)), nil
}
