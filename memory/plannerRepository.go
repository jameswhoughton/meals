package memory

import (
	"fmt"
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
	Meals   []planner.Meal
	Planner map[string]int
}

func (pr *PlannerRepository) findMeal(id int) (planner.Meal, error) {
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

func (pr *PlannerRepository) Add(date time.Time, mealId int) error {
	meal, err := pr.findMeal(mealId)

	if err != nil {
		return err
	}

	key := date.Format("2006-01-02") + "|" + strconv.Itoa(meal.UserId)

	pr.Planner[key] = meal.Id

	return nil
}

func (pr *PlannerRepository) Get(date time.Time, userId int) (planner.Meal, error) {
	key := date.Format("2006-01-02") + "|" + strconv.Itoa(userId)
	mealId := pr.Planner[key]

	meal, err := pr.findMeal(mealId)

	if err != nil {
		return meal, err
	}

	return meal, nil
}

func (pr *PlannerRepository) Clear(date time.Time, userId int) error {
	key := date.Format("2006-01-02") + "|" + strconv.Itoa(userId)

	delete(pr.Planner, key)

	return nil
}
