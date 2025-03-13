package memory

import (
	"strings"

	"github.com/jameswhoughton/meals/internal/meals"
)

type IngredientRepository struct {
	Store []meals.Ingredient
}

func (i *IngredientRepository) Get(id int) (meals.Ingredient, error) {
	for _, ingredient := range i.Store {
		if ingredient.Id == id {
			return ingredient, nil
		}
	}

	return meals.Ingredient{}, meals.ErrorIngredientNotFound{Id: id}
}

func (i *IngredientRepository) List(filter meals.IngredientFilter) ([]meals.Ingredient, error) {
	err := filter.Validate()

	if err != nil {
		return []meals.Ingredient{}, err
	}

	var results []meals.Ingredient

	for _, ingredient := range i.Store {
		if ingredient.UserId != filter.UserId {
			continue
		}

		if filter.Name == nil {
			results = append(results, ingredient)
			continue
		}

		if strings.Contains(strings.ToLower(ingredient.Name), strings.ToLower(*filter.Name)) {
			results = append(results, ingredient)
		}
	}

	return results, nil
}

func (i *IngredientRepository) Create(ingredient meals.Ingredient) (meals.Ingredient, error) {
	ingredient.Id = len(i.Store) + 1

	i.Store = append(i.Store, ingredient)

	return ingredient, nil
}

func (i *IngredientRepository) Update(ingredient meals.Ingredient) error {
	for n, storedIngredient := range i.Store {
		if storedIngredient.Id == ingredient.Id {
			i.Store[n] = ingredient
		}
	}

	return nil
}

func (i *IngredientRepository) Destroy(id int) error {
	var updatedStore []meals.Ingredient

	for _, ingredient := range i.Store {
		if ingredient.Id == id {
			continue
		}

		updatedStore = append(updatedStore, ingredient)
	}

	i.Store = updatedStore

	return nil
}
