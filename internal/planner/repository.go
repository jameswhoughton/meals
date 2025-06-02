package planner

import (
	"context"
	"time"
)

type Meal struct {
	Id     int
	UserId int
	Name   string
}

type Ingredient struct {
	Name     string
	Quantity int
	Unit     string
}

type Repository interface {
	Get(ctx context.Context, date time.Time, userId int) (Meal, error)
	Add(ctx context.Context, date time.Time, mealId int) error
	Clear(ctx context.Context, date time.Time, userId int) error
	GetIngredients(ctx context.Context, startDate, endDate time.Time, userId int) ([]Ingredient, error)
}
