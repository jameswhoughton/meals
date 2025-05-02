package planner

import "time"

type Meal struct {
	Id     int
	UserId int
	Name   string
}

type Repository interface {
	Get(date time.Time, userId int) (Meal, error)
	Add(date time.Time, mealId int) error
	Clear(date time.Time, userId int) error
}
