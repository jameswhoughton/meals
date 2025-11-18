package meals

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"time"
)

type PlannerRepository interface {

	// Returns the ids of meals for the given date range and userId.
	// Grouped by date
	//
	// The time portion of the date is ignored.
	GetMealIdsInRange(ctx context.Context, startDate, endDate time.Time, userId int) (map[DateKey][]int, error)

	// Assign a mealId to the given date
	//
	// The time portion of the date is ignored.
	Add(ctx context.Context, date time.Time, mealId int) error

	// Remove any meals assigned to the date by the userID
	//
	// The time portion of the date is ignored.
	Clear(ctx context.Context, date time.Time, userId int) error
}

var ErrValidation = errors.New("Parameters invalid")

type DateKey string

func GetDateKey(d time.Time) DateKey {
	return DateKey(d.Format("2006-01-02"))
}

type Day struct {
	Date      string
	Label     string
	Meals     []Meal
	IsWeekend bool
	IsToday   bool
}

type IngredientTotal struct {
	Name     string
	Quantity int
	Unit     string
}

func NewPlannerService(plannerRepo PlannerRepository, mealRepo MealRepository, mealMetaRepo MealMetaDataRepository) *PlannerService {
	return &PlannerService{
		plannerRepo:  plannerRepo,
		mealRepo:     mealRepo,
		mealMetaRepo: mealMetaRepo,
	}
}

type PlannerService struct {
	plannerRepo  PlannerRepository
	mealRepo     MealRepository
	mealMetaRepo MealMetaDataRepository
}

// Get the meals assigned to the range of dates
func (s *PlannerService) GetMeals(ctx context.Context, startDate time.Time, numberOfDays, userId int) ([]Day, error) {
	if numberOfDays < 1 {
		return nil, errors.New("PlannerService.GetMeals: numberOfDays must be 1 or more")
	}

	days := make([]Day, numberOfDays)
	date := startDate
	endDate := startDate.AddDate(0, 0, numberOfDays)

	mealIds, err := s.plannerRepo.GetMealIdsInRange(ctx, startDate, endDate, userId)

	if err != nil {
		return days, fmt.Errorf("PlannerPlannerService.GetMeals: error fetching assigned meals: %v", err)
	}

	for i := range days {
		assignedMealIds := mealIds[GetDateKey(date)]
		assignedMeals := make([]Meal, len(assignedMealIds))

		for i, assignedMealId := range assignedMealIds {
			meal, err := s.mealRepo.Get(ctx, assignedMealId)

			if err != nil {
				return days, fmt.Errorf("PlannerPlannerService.GetMeals: error fetching meal id=%d: %v", assignedMealId, err)
			}

			assignedMeals[i] = meal
		}

		day := Day{
			Date:      date.Format("2006-01-02"),
			Label:     fmt.Sprintf("%s (%s)", date.Weekday().String(), date.Format("02/01")),
			Meals:     assignedMeals,
			IsWeekend: slices.Contains([]string{time.Saturday.String(), time.Sunday.String()}, date.Weekday().String()),
			IsToday:   date.Format("2006-01-02") == time.Now().Format("2006-01-02"),
		}

		days[i] = day

		date = date.AddDate(0, 0, 1)
	}

	return days, nil
}

func (s *PlannerService) GetMealIdsForDate(ctx context.Context, date time.Time, userId int) ([]int, error) {
	days, err := s.plannerRepo.GetMealIdsInRange(ctx, date, date, userId)

	if err != nil {
		return nil, err
	}

	return days[GetDateKey(date)], nil
}

func (s *PlannerService) GetIngredients(ctx context.Context, startDate, endDate time.Time, userId int) ([]IngredientTotal, error) {
	if endDate.Before(startDate) {
		return []IngredientTotal{}, ErrValidation
	}

	maxDate := startDate.AddDate(0, 0, 28)

	if maxDate.Before(endDate) {
		return []IngredientTotal{}, ErrValidation
	}

	mealIdByDate, err := s.plannerRepo.GetMealIdsInRange(ctx, startDate, endDate, userId)

	if err != nil {
		return []IngredientTotal{}, fmt.Errorf("PlannerPlannerService.GetIngredients: error fetching assigned meals: %v", err)
	}

	mealIds := slices.Concat(slices.Collect(maps.Values(mealIdByDate))...)

	totals, err := s.mealMetaRepo.GetTotalIngredients(ctx, mealIds)

	if err != nil {
		return []IngredientTotal{}, fmt.Errorf("PlannerPlannerService.GetIngredients: error fetching total ingredients: %v", err)
	}

	return totals, nil
}
