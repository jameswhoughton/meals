package meals

import (
	"fmt"
)

type Service struct {
	repo Repository
}

type ErrorFormInvalid struct{}

func (e ErrorFormInvalid) Error() string {
	return "Meal form invalid"
}

func (s *Service) CreateMeal(meal *Meal) (Meal, error) {
	if isValid := meal.Validate(); !isValid {
		return Meal{}, ErrorFormInvalid{}
	}

	createdMeal, err := s.repo.Create(*meal)

	if err != nil {
		return Meal{}, fmt.Errorf("Error creating meal: %v", err)
	}

	return createdMeal, nil
}

func (s *Service) UpdateMeal(meal *Meal) error {
	if isValid := meal.Validate(); !isValid {
		return ErrorFormInvalid{}
	}

	err := s.repo.Update(*meal)

	if err != nil {
		return fmt.Errorf("Error updating meal: %v", err)
	}

	return nil
}

func (s *Service) DestroyMeal(id int) error {
	return s.repo.Destroy(id)
}

func (s *Service) ListMeals(filter MealFilter) ([]Meal, error) {
	return s.repo.List(filter)
}

// Assign a meal to a date

// Return 1 or more random meals restricted by parameters/rules
