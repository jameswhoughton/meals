package meals

import (
	"context"
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

func (s *Service) populateIngredientIds(ctx context.Context, meal *Meal) error {
	ingredientNames := make([]string, len(meal.Ingredients))

	for i, ingredient := range meal.Ingredients {
		if ingredient.Id > 0 {
			continue
		}

		ingredientNames[i] = ingredient.Name
	}

	ingredientIds, err := s.ingredients.FromNames(ctx, ingredientNames, meal.UserId)

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

func (s *Service) populateTagIds(ctx context.Context, meal *Meal) error {
	tagNames := make([]string, len(meal.Tags))

	for i, tag := range meal.Tags {
		if tag.Id > 0 {
			continue
		}

		tagNames[i] = tag.Name
	}

	tagIds, err := s.tags.FromNames(ctx, tagNames, meal.UserId)

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

func (s *Service) CreateMeal(ctx context.Context, meal *Meal) (Meal, error) {
	if isValid := meal.Validate(); !isValid {
		return Meal{}, ErrorFormInvalid{}
	}

	meal.CreatedAt = time.Now()
	meal.UpdatedAt = time.Now()

	s.populateIngredientIds(ctx, meal)
	s.populateTagIds(ctx, meal)

	createdMeal, err := s.meals.Create(ctx, *meal)

	if err != nil {
		return Meal{}, fmt.Errorf("Error creating meal: %v", err)
	}

	return createdMeal, nil
}

func (s *Service) UpdateMeal(ctx context.Context, meal *Meal) error {
	if isValid := meal.Validate(); !isValid {
		return ErrorFormInvalid{}
	}

	meal.UpdatedAt = time.Now()

	s.populateIngredientIds(ctx, meal)
	s.populateTagIds(ctx, meal)

	err := s.meals.Update(ctx, *meal)

	if err != nil {
		return fmt.Errorf("Error updating meal: %v", err)
	}

	return nil
}

func (s *Service) UpdateIngredient(ctx context.Context, ingredient *Ingredient) error {
	if isValid := ingredient.Validate(); !isValid {
		return ErrorFormInvalid{}
	}

	err := s.ingredients.Update(ctx, *ingredient)

	if err != nil {
		return fmt.Errorf("Error updating ingredient: %v", err)
	}

	return nil
}

func (s *Service) UpdateTag(ctx context.Context, tag *Tag) error {
	if isValid := tag.Validate(); !isValid {
		return ErrorFormInvalid{}
	}

	err := s.tags.Update(ctx, *tag)

	if err != nil {
		return fmt.Errorf("Error updating tag: %v", err)
	}

	return nil
}
