package memory

import (
	"context"
	"maps"
	"slices"
	"strconv"
	"time"

	"github.com/jameswhoughton/meals"
)

func NewPlannerRepository() *PlannerRepository {
	store := make(map[string][]int)

	return &PlannerRepository{
		Planner: store,
	}
}

type PlannerRepository struct {
	Meals   []meals.Meal
	Planner map[string][]int
}

func plannerKey(d time.Time, userId int) string {
	return string(meals.GetDateKey(d)) + "|" + strconv.Itoa(userId)
}

func (pr *PlannerRepository) findMeal(_ context.Context, id int) (meals.Meal, error) {
	var meal meals.Meal

	for _, m := range pr.Meals {
		if m.Id == id {
			meal = m
			break
		}
	}

	if meal.Id == 0 {
		return meal, meals.ErrMealNotFound
	}

	return meal, nil
}

func (pr *PlannerRepository) Add(ctx context.Context, date time.Time, mealId int) error {
	meal, err := pr.findMeal(ctx, mealId)

	if err != nil {
		return err
	}

	key := plannerKey(date, meal.UserId)

	if _, ok := pr.Planner[key]; !ok {
		pr.Planner[key] = []int{}
	}

	pr.Planner[key] = append(pr.Planner[key], meal.Id)

	return nil
}

func (pr *PlannerRepository) GetMealIdsInRange(ctx context.Context, startDate, endDate time.Time, userId int) (map[meals.DateKey][]int, error) {
	date := startDate
	diff := int(endDate.AddDate(0, 0, 1).Sub(startDate).Hours() / 24)
	mealIds := make(map[meals.DateKey][]int, diff)

	for date.Before(endDate.AddDate(0, 0, 1)) {
		internalKey := plannerKey(date, userId)
		plannedIds, ok := pr.Planner[internalKey]

		dateKey := meals.GetDateKey(date)
		mealIds[dateKey] = []int{}

		date = date.AddDate(0, 0, 1)

		if !ok {
			continue
		}

		mealIds[dateKey] = plannedIds
	}

	return mealIds, nil
}

func (pr *PlannerRepository) GetMealIdsForDate(ctx context.Context, date time.Time, userId int) ([]int, error) {
	days, err := pr.GetMealIdsInRange(ctx, date, date, userId)

	if err != nil {
		return nil, err
	}

	return days[meals.GetDateKey(date)], nil
}

func (pr *PlannerRepository) Clear(ctx context.Context, date time.Time, userId int) error {
	key := plannerKey(date, userId)

	delete(pr.Planner, key)

	return nil
}

func (pr *PlannerRepository) GetIngredientsInRange(ctx context.Context, startDate, endDate time.Time, userId int) ([]meals.IngredientTotal, error) {
	key := plannerKey(startDate, userId)

	totals := make(map[string]meals.IngredientTotal)

	for startDate.Before(endDate.AddDate(0, 0, 1)) {
		plannedMealIds, ok := pr.Planner[key]
		startDate = startDate.AddDate(0, 0, 1)
		key = plannerKey(startDate, userId)

		if !ok {
			continue
		}

		for _, mealId := range plannedMealIds {
			meal, err := pr.findMeal(ctx, mealId)

			if err != nil {
				return nil, err
			}

			ingredients := meal.Ingredients

			for _, ingredient := range ingredients {
				totalsKey := ingredient.Name + "|" + ingredient.Unit

				total, ok := totals[totalsKey]

				if !ok {
					totals[totalsKey] = meals.IngredientTotal{
						Name:     ingredient.Name,
						Quantity: ingredient.Quantity,
						Unit:     ingredient.Unit,
					}

					continue
				}

				total.Quantity += ingredient.Quantity

				totals[totalsKey] = total
			}
		}

	}

	return slices.Collect(maps.Values(totals)), nil
}
