package meals

import (
	"fmt"
	"time"
)

type Service struct {
	meals       MealRepository
	ingredients IngredientRepository
	tags        TagRepository
}

type ErrorFormInvalid struct{}

func (e ErrorFormInvalid) Error() string {
	return "Form invalid"
}

func NewService(m MealRepository, i IngredientRepository, t TagRepository) Service {
	return Service{m, i, t}
}

func (s *Service) CreateMeal(meal *Meal) (Meal, error) {
	if isValid := meal.Validate(); !isValid {
		return Meal{}, ErrorFormInvalid{}
	}

	meal.CreatedAt = time.Now()
	meal.UpdatedAt = time.Now()

	createdMeal, err := s.meals.Create(*meal)

	if err != nil {
		return Meal{}, fmt.Errorf("Error creating meal: %v", err)
	}

	return createdMeal, nil
}

func (s *Service) UpdateMeal(meal *Meal) error {
	if isValid := meal.Validate(); !isValid {
		return ErrorFormInvalid{}
	}

	meal.UpdatedAt = time.Now()

	err := s.meals.Update(*meal)

	if err != nil {
		return fmt.Errorf("Error updating meal: %v", err)
	}

	return nil
}

func (s *Service) UpdateIngredient(ingredient *Ingredient) error {
	if isValid := ingredient.Validate(); !isValid {
		return ErrorFormInvalid{}
	}

	err := s.ingredients.Update(*ingredient)

	if err != nil {
		return fmt.Errorf("Error updating ingredient: %v", err)
	}

	return nil
}

func (s *Service) UpdateTag(tag *Tag) error {
	if isValid := tag.Validate(); !isValid {
		return ErrorFormInvalid{}
	}

	err := s.tags.Update(*tag)

	if err != nil {
		return fmt.Errorf("Error updating tag: %v", err)
	}

	return nil
}

// Return 1 or more random meals restricted by parameters/rules
