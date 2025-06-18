package planner

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"
)

var ErrValidation = errors.New("Parameters invalid")

type Day struct {
	Date      string
	Label     string
	Meal      Meal
	IsWeekend bool
	IsToday   bool
}

func NewService(repo Repository) *Service {
	return &Service{repo}
}

type Service struct {
	repo Repository
}

// Get the meals assigned to the range of dates
func (s *Service) GetMeals(ctx context.Context, startDate time.Time, numberOfDays, userId int) ([]Day, error) {
	days := make([]Day, numberOfDays)
	date := startDate

	for i := range days {
		meal, err := s.repo.Get(ctx, date, userId)

		if err != nil && !errors.Is(err, ErrMealNotSet) {
			return days, fmt.Errorf("PlannerService.GetMeals: Error fetching meal: %v", err)
		}

		day := Day{
			Date:      date.Format("2006-01-02"),
			Label:     fmt.Sprintf("%s (%s)", date.Weekday().String(), date.Format("02/01")),
			Meal:      meal,
			IsWeekend: slices.Contains([]string{time.Saturday.String(), time.Sunday.String()}, date.Weekday().String()),
			IsToday:   date.Format("2006-01-02") == time.Now().Format("2006-01-02"),
		}

		days[i] = day

		date = date.Add(24 * time.Hour)
	}

	return days, nil
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
