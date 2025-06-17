package meals

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Service struct {
	meals Repository
}

var ErrMealFormInvalid = errors.New("meal form has validation errors")
var ErrMealFilterFormInvalid = errors.New("filter form has validation errors")

type MealFilterForm struct {
	UserId int
	Name   *string
	Tags   []string
	Errors map[string]string
}

func (mf *MealFilterForm) Validate() bool {
	mf.Errors = make(map[string]string)

	if mf.UserId == 0 {
		mf.Errors["user_id"] = "UserID must be set"
	}

	return len(mf.Errors) == 0
}

func NewService(m Repository) Service {
	return Service{m}
}

func (s *Service) FilterMeals(ctx context.Context, form *MealFilterForm) ([]Meal, error) {
	if !form.Validate() {
		return []Meal{}, ErrMealFilterFormInvalid
	}

	filter := MealFilter{
		UserId: form.UserId,
		Name:   form.Name,
		Tags:   form.Tags,
	}

	return s.meals.Find(ctx, filter)
}

func (s *Service) CreateMeal(ctx context.Context, meal *Meal) (Meal, error) {
	if isValid := meal.Validate(); !isValid {
		return Meal{}, ErrMealFormInvalid
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
		return ErrMealFormInvalid
	}

	meal.UpdatedAt = time.Now()

	err := s.meals.Update(ctx, *meal)

	if err != nil {
		return fmt.Errorf("Error updating meal: %v", err)
	}

	return nil
}
