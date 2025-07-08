package memory

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"time"

	"github.com/jameswhoughton/meals"
)

func NewPlannerRepository() *PlannerRepository {
	store := make(map[string]int)

	return &PlannerRepository{
		Planner: store,
	}
}

type PlannerRepository struct {
	Meals   []meals.Meal
	Planner map[string]int
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

	key := date.Format("2006-01-02") + "|" + strconv.Itoa(meal.UserId)

	pr.Planner[key] = meal.Id

	return nil
}

func (pr *PlannerRepository) GetMealIdsInRange(ctx context.Context, startDate, endDate time.Time, userId int) (map[int]int, error) {
	date := startDate
	diff := int(endDate.AddDate(0, 0, 1).Sub(startDate).Hours() / 24)
	mealIds := make(map[int]int, diff)

	for date.Before(endDate.AddDate(0, 0, 1)) {
		key := date.Format("2006-01-02") + "|" + strconv.Itoa(userId)
		fmt.Println(key)
		mealId, ok := pr.Planner[key]

		weekDay := int(date.Weekday())
		mealIds[weekDay] = 0

		date = date.AddDate(0, 0, 1)

		if !ok {
			continue
		}

		mealIds[weekDay] = mealId
	}

	fmt.Println("MEALIDS", mealIds, diff, startDate, endDate)

	return mealIds, nil
}

func (pr *PlannerRepository) Clear(ctx context.Context, date time.Time, userId int) error {
	key := date.Format("2006-01-02") + "|" + strconv.Itoa(userId)

	delete(pr.Planner, key)

	return nil
}

func (pr *PlannerRepository) GetIngredientsInRange(ctx context.Context, startDate, endDate time.Time, userId int) ([]meals.IngredientTotal, error) {
	key := startDate.Format("2006-01-02") + "|" + strconv.Itoa(userId)

	totals := make(map[string]meals.IngredientTotal)

	for startDate.Before(endDate.AddDate(0, 0, 1)) {
		mealId, ok := pr.Planner[key]
		startDate = startDate.AddDate(0, 0, 1)
		key = startDate.Format("2006-01-02") + "|" + strconv.Itoa(userId)

		if !ok {
			continue
		}

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

	return slices.Collect(maps.Values(totals)), nil
}
