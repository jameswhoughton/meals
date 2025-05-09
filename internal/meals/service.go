package meals

import (
	"context"
	"fmt"
	"time"
)

type Service struct {
	meals Repository
}

type ErrorFormInvalid struct{}

func (e ErrorFormInvalid) Error() string {
	return "Form invalid"
}

func NewService(m Repository) Service {
	return Service{m}
}

func (s *Service) CreateMeal(ctx context.Context, meal *Meal) (Meal, error) {
	if isValid := meal.Validate(); !isValid {
		return Meal{}, ErrorFormInvalid{}
	}

	meal.CreatedAt = time.Now()
	meal.UpdatedAt = time.Now()

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

	err := s.meals.Update(ctx, *meal)

	if err != nil {
		return fmt.Errorf("Error updating meal: %v", err)
	}

	return nil
}
