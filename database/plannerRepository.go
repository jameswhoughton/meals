package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jameswhoughton/meals/internal/planner"
)

func NewPlannerRepository(db *sql.DB) *PlannerRepository {
	return &PlannerRepository{
		db: db,
	}
}

type PlannerRepository struct {
	db *sql.DB
}

func normaliseDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
}

func (pr *PlannerRepository) Add(ctx context.Context, date time.Time, mealId int) error {
	_, err := pr.db.ExecContext(ctx, `
		INSERT INTO planner 
		(meal_id, meal_date)
		VALUES (?, ?)
	`, mealId, normaliseDate(date))

	if err != nil {
		return err
	}

	return nil
}

func (pr *PlannerRepository) Get(ctx context.Context, date time.Time, userId int) (planner.Meal, error) {
	var meal planner.Meal

	err := pr.db.QueryRowContext(ctx, `
		SELECT m.id, m.user_id, m.name 
		FROM planner p 
		LEFT JOIN meals m 
		ON p.meal_id = m.id
		WHERE p.meal_date = ?
		AND m.user_id = ?
		AND m.id IS NOT NULL
	`, normaliseDate(date), userId).Scan(&meal.Id, &meal.UserId, &meal.Name)

	if err != nil {
		if err == sql.ErrNoRows {
			return meal, planner.ErrMealNotSet
		}

		return meal, fmt.Errorf("PlannerRepository: Unable to fetch meal for date: %v", err)
	}

	return meal, nil
}

func (pr *PlannerRepository) Clear(ctx context.Context, date time.Time, userId int) error {
	_, err := pr.db.ExecContext(ctx, `
		DELETE p FROM planner p
		LEFT JOIN meals m
		ON m.id = p.meal_id
		WHERE meal_date = ?
		AND user_id = ?
	`, normaliseDate(date), userId)

	if err != nil {
		return fmt.Errorf("Planner Repository: Unable to clear date %s for user %d: %v", date, userId, err)
	}

	return nil
}

func (pr *PlannerRepository) GetIngredients(ctx context.Context, startDate, endDate time.Time, userId int) ([]planner.Ingredient, error) {
	var totals []planner.Ingredient

	rows, err := pr.db.QueryContext(ctx, `
		SELECT i.name, SUM(i.quantity), i.unit
		FROM planner p
		LEFT JOIN meal_ingredients i
		ON p.meal_id = i.meal_id
		LEFT JOIN meals m
		ON p.meal_id = m.id
		WHERE p.meal_date >= ?
		AND p.meal_date <= ?
		AND m.user_id = ?
		GROUP BY name, unit
	`, startDate, endDate, userId)

	if err != nil {
		if err == sql.ErrNoRows {
			return totals, nil
		}

		return totals, fmt.Errorf("Planner Repository: unable to get ingredients: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var ingredient planner.Ingredient

		err := rows.Scan(&ingredient.Name, &ingredient.Quantity, &ingredient.Unit)

		if err != nil {
			return totals, fmt.Errorf("Planner Repository: Unable to scan row: %v", err)
		}

		totals = append(totals, ingredient)
	}

	return totals, nil
}
