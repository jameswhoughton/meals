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

	// Returns the ids of meals the given date range for the given userId.
	//
	// An id of 0 indicates that no meal is set
	// The time portion of the date is ignored.
	GetMealIdsInRange(ctx context.Context, startDate, endDate time.Time, userId int) (map[int]int, error)

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

type Day struct {
	Date      string
	Label     string
	Meal      Meal
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
	days := make([]Day, numberOfDays)
	date := startDate
	endDate := startDate.AddDate(0, 0, 7)

	mealIds, err := s.plannerRepo.GetMealIdsInRange(ctx, startDate, endDate, userId)

	if err != nil {
		return days, fmt.Errorf("PlannerPlannerService.GetMeals: error fetching assigned meals: %v", err)
	}

	for i := range days {
		var meal Meal

		assignedMealId := mealIds[int(date.Weekday())]

		if assignedMealId > 0 {
			meal, err = s.mealRepo.Get(ctx, assignedMealId)

			if err != nil {
				return days, fmt.Errorf("PlannerPlannerService.GetMeals: error fetching meal id=%d: %v", assignedMealId, err)
			}
		}

		day := Day{
			Date:      date.Format("2006-01-02"),
			Label:     fmt.Sprintf("%s (%s)", date.Weekday().String(), date.Format("02/01")),
			Meal:      meal,
			IsWeekend: slices.Contains([]string{time.Saturday.String(), time.Sunday.String()}, date.Weekday().String()),
			IsToday:   date.Format("2006-01-02") == time.Now().Format("2006-01-02"),
		}

		days[i] = day

		date = date.AddDate(0, 0, 1)
	}

	return days, nil
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

	mealIds := slices.Collect(maps.Values(mealIdByDate))

	totals, err := s.mealMetaRepo.GetTotalIngredients(ctx, mealIds)

	if err != nil {
		return []IngredientTotal{}, fmt.Errorf("PlannerPlannerService.GetIngredients: error fetching total ingredients: %v", err)
	}

	return totals, nil
}
