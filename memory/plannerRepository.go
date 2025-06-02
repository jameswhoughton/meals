package memory

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"time"

	"github.com/jameswhoughton/meals/internal/planner"
)

func NewPlannerRepository() *PlannerRepository {
	store := make(map[string]int)

	return &PlannerRepository{
		Planner: store,
	}
}

type PlannerRepository struct {
	Meals       []planner.Meal
	Ingredients map[int][]planner.Ingredient
	Planner     map[string]int
}

func (pr *PlannerRepository) findMeal(_ context.Context, id int) (planner.Meal, error) {
	var meal planner.Meal

	for _, m := range pr.Meals {
		if m.Id == id {
			meal = m
			break
		}
	}

	if meal.Id == 0 {
		return meal, fmt.Errorf("Meal with the id %d not found", id)
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

func (pr *PlannerRepository) Get(ctx context.Context, date time.Time, userId int) (planner.Meal, error) {
	key := date.Format("2006-01-02") + "|" + strconv.Itoa(userId)
	mealId := pr.Planner[key]

	meal, err := pr.findMeal(ctx, mealId)

	if err != nil {
		return meal, err
	}

	return meal, nil
}

func (pr *PlannerRepository) Clear(ctx context.Context, date time.Time, userId int) error {
	key := date.Format("2006-01-02") + "|" + strconv.Itoa(userId)

	delete(pr.Planner, key)

	return nil
}

func (pr *PlannerRepository) GetIngredients(ctx context.Context, startDate, endDate time.Time, userId int) ([]planner.Ingredient, error) {
	key := startDate.Format("2006-01-02") + "|" + strconv.Itoa(userId)

	totals := make(map[string]planner.Ingredient)

	for !startDate.Equal(endDate.Add(24 * time.Hour)) {
		mealId, ok := pr.Planner[key]
		startDate = startDate.Add(24 * time.Hour)
		key = startDate.Format("2006-01-02") + "|" + strconv.Itoa(userId)

		if !ok {
			continue
		}

		ingredients := pr.Ingredients[mealId]

		for _, ingredient := range ingredients {
			totalsKey := ingredient.Name + "|" + ingredient.Unit

			total, ok := totals[totalsKey]

			if !ok {
				totals[totalsKey] = ingredient

				continue
			}

			total.Quantity += ingredient.Quantity

			totals[totalsKey] = total
		}

	}
	fmt.Println(totals, pr.Planner)

	return slices.Collect(maps.Values(totals)), nil
}
