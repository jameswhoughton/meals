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

func (s *Service) populateIngredientIds(meal *Meal) error {
	ingredientNames := make([]string, len(meal.Ingredients))

	for i, ingredient := range meal.Ingredients {
		if ingredient.Id > 0 {
			continue
		}

		ingredientNames[i] = ingredient.Name
	}

	ingredientIds, err := s.ingredients.FromNames(ingredientNames, meal.UserId)

	if err != nil {
		return err
	}

	for i, name := range ingredientNames {
		if ingredientIds[name] == 0 {
			continue
		}

		meal.Ingredients[i].Id = ingredientIds[name]
	}

	return nil
}

func (s *Service) populateTagIds(meal *Meal) error {
	tagNames := make([]string, len(meal.Tags))

	for i, tag := range meal.Tags {
		if tag.Id > 0 {
			continue
		}

		tagNames[i] = tag.Name
	}

	tagIds, err := s.tags.FromNames(tagNames, meal.UserId)

	if err != nil {
		return err
	}

	for i, name := range tagNames {
		if tagIds[name] == 0 {
			continue
		}

		meal.Tags[i].Id = tagIds[name]
	}

	return nil
}

func (s *Service) CreateMeal(meal *Meal) (Meal, error) {
	if isValid := meal.Validate(); !isValid {
		return Meal{}, ErrorFormInvalid{}
	}

	meal.CreatedAt = time.Now()
	meal.UpdatedAt = time.Now()

	s.populateIngredientIds(meal)
	s.populateTagIds(meal)

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

	s.populateIngredientIds(meal)
	s.populateTagIds(meal)

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
