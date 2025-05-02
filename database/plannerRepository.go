package database

import (
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

func (pr *PlannerRepository) Add(date time.Time, mealId int) error {
	_, err := pr.db.Exec(`
		INSERT INTO planner 
		(meal_id, date)
		VALUES (?, ?)
	`, mealId, normaliseDate(date))

	if err != nil {
		return err
	}

	return nil
}

func (pr *PlannerRepository) Get(date time.Time, userId int) (planner.Meal, error) {
	var meal planner.Meal

	err := pr.db.QueryRow(`
		SELECT m.id, m.user_id, m.name 
		FROM planner p 
		LEFT JOIN meals m 
		ON p.meal_id = m.id
		WHERE p.date = ?
		AND m.user_id = ?
		AND m.id IS NOT NULL
	`, normaliseDate(date), userId).Scan(&meal.Id, &meal.UserId, &meal.Name)

	if err != nil {
		if err == sql.ErrNoRows {
			return meal, fmt.Errorf("Planner Repository: No meal set for user %d on date %s", userId, date)
		}

		return meal, fmt.Errorf("PlannerRepository: Unable to fetch meal for date: %v", err)
	}

	return meal, nil
}

func (pr *PlannerRepository) Clear(date time.Time, userId int) error {
	_, err := pr.db.Exec(`
		DELETE FROM planner
		WHERE id IN (
			SELECT p.id FROM planner p
			LEFT JOIN meals m
			ON m.id = p.meal_id
			WHERE date = ?
			AND user_id = ?
		)
	`, normaliseDate(date), userId)

	if err != nil {
		return fmt.Errorf("Planner Repository: Unable to clear date %s for user %d: %v", date, userId, err)
	}

	return nil
}
