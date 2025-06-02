package planner

import (
	"context"
	"errors"
	"time"
)

var ErrValidation = errors.New("Parameters invalid")

func NewService(repo Repository) *Service {
	return &Service{repo}
}

type Service struct {
	repo Repository
}

func (s *Service) GetIngredients(ctx context.Context, startDate, endDate time.Time, userId int) ([]Ingredient, error) {
	if endDate.Before(startDate) {
		return []Ingredient{}, ErrValidation
	}

	maxDate := startDate.AddDate(0, 0, 28)

	if maxDate.Before(endDate) {
		return []Ingredient{}, ErrValidation
	}

	return s.repo.GetIngredients(ctx, startDate, endDate, userId)
}
