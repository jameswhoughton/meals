package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jameswhoughton/meals"
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
		return fmt.Errorf("PlannerRepository.Add: error adding record: %v", err)
	}

	return nil
}

func (pr *PlannerRepository) GetMealIdsInRange(ctx context.Context, startDate, endDate time.Time, userId int) (map[meals.DateKey][]int, error) {
	startDate = normaliseDate(startDate)
	endDate = normaliseDate(endDate)
	diff := int(startDate.Sub(endDate).Hours() / 24)
	mealIds := make(map[meals.DateKey][]int, diff)

	// Populate map
	date := startDate

	for date.Before(endDate.AddDate(0, 0, 1)) {
		mealIds[meals.GetDateKey(date)] = []int{}

		date = date.AddDate(0, 0, 1)
	}

	rows, err := pr.db.QueryContext(ctx, `
		SELECT m.id, p.meal_date
		FROM planner p 
		LEFT JOIN meals m 
		ON p.meal_id = m.id
		WHERE p.meal_date >= ?
		AND p.meal_date <= ?
		AND m.user_id = ?
		AND m.id IS NOT NULL
	`, normaliseDate(startDate), normaliseDate(endDate), userId)

	if err != nil {
		return mealIds, fmt.Errorf("PlannerRepository.GetMealIdsInRange: Unable to fetch meal IDs in range startDate=%v endDate=%v: %v", startDate, endDate, err)
	}

	defer rows.Close()

	for rows.Next() {
		var (
			date   time.Time
			mealId int
		)

		err := rows.Scan(&mealId, &date)

		if err != nil {
			return mealIds, fmt.Errorf("PlannerRepository.GetMealIdsInRange: Unable to scan row: %v", err)
		}

		mealIds[meals.GetDateKey(date)] = append(mealIds[meals.GetDateKey(date)], mealId)
	}

	return mealIds, nil
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
		return fmt.Errorf("PlannerRepository: Unable to clear date %s for user %d: %v", date, userId, err)
	}

	return nil
}
