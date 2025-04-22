package memory

import (
	"strings"

	"github.com/jameswhoughton/meals/internal/meals"
)

type IngredientRepository struct {
	Store []meals.Ingredient
}

func (ir *IngredientRepository) GetById(id int) (meals.Ingredient, error) {
	for _, ingredient := range ir.Store {
		if ingredient.Id == id {
			return ingredient, nil
		}
	}

	return meals.Ingredient{}, meals.ErrorIngredientNotFound{Id: id}
}

func (ir *IngredientRepository) Find(search string, userId int) ([]meals.Ingredient, error) {
	var ingredients []meals.Ingredient

	search = strings.ToLower(search)

	for _, ingredient := range ir.Store {
		if ingredient.UserId != userId {
			continue
		}

		if strings.Contains(strings.ToLower(ingredient.Name), search) {
			ingredients = append(ingredients, ingredient)
		}
	}

	return ingredients, nil
}

func (ir *IngredientRepository) Update(ingredient meals.Ingredient) error {
	for i, existingIngredient := range ir.Store {
		if ingredient.Id == existingIngredient.Id {
			ir.Store[i].Name = ingredient.Name
			break
		}
	}

	return nil
}

func (ir *IngredientRepository) FromNames(names []string, userId int) (map[string]int, error) {
	ingredientMap := make(map[string]int, len(names))

	for _, ingredient := range ir.Store {
		if ingredient.UserId != userId {
			continue
		}

		ingredientMap[ingredient.Name] = ingredient.Id
	}

	return ingredientMap, nil
}
