package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jameswhoughton/meals"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRespository(db *sql.DB) *UserRepository {
	return &UserRepository{db}
}

func (us *UserRepository) GetById(ctx context.Context, id int) (meals.User, error) {
	var user meals.User

	err := us.db.QueryRowContext(
		ctx,
		`SELECT id, name, email, meal_start_day, password FROM users WHERE id = ?`,
		id).Scan(&user.Id, &user.Name, &user.Email, &user.MealStartDay, &user.Password)

	if err != nil {
		if err == sql.ErrNoRows {
			return meals.User{}, meals.ErrUserNotFound
		}

		return meals.User{}, fmt.Errorf("failed to query user with id=%d: %w", id, err)
	}

	return user, nil
}

func (us *UserRepository) GetByEmail(ctx context.Context, email string) (meals.User, error) {
	var user meals.User

	err := us.db.QueryRowContext(
		ctx,
		`SELECT id, name, email, meal_start_day, password FROM users WHERE email = ?`,
		email).Scan(&user.Id, &user.Name, &user.Email, &user.MealStartDay, &user.Password)

	if err != nil {
		if err == sql.ErrNoRows {
			return meals.User{}, meals.ErrUserNotFound
		}

		return meals.User{}, fmt.Errorf("failed to query user with email=%s: %w", email, err)
	}

	return user, nil
}

func (us *UserRepository) Create(ctx context.Context, user meals.User) (meals.User, error) {
	result, err := us.db.ExecContext(ctx, `
		INSERT INTO users 
		(email, name, password, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, user.Email, user.Name, user.Password, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		return meals.User{}, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return meals.User{}, err
	}

	user.Id = int(id)

	return user, nil
}

func (us *UserRepository) Update(ctx context.Context, user meals.UserUpdate) error {
	_, err := us.GetById(ctx, user.Id)

	if err != nil {
		return err
	}

	updates := []string{
		"updated_at = ?",
		"email = ?",
		"name = ?",
		"meal_start_day = ?",
	}

	values := []any{
		time.Now(),
		user.Email,
		user.Name,
		user.MealStartDay,
	}

	if v := user.Password; v != nil {
		updates = append(updates, "password = ?")
		values = append(values, *v)
	}

	values = append(values, user.Id)

	query := "UPDATE users SET " + strings.Join(updates, ",") + " WHERE id = ?"

	_, err = us.db.ExecContext(ctx, query, values...)

	if err != nil {
		return fmt.Errorf("error updating user: %v", err)
	}

	return nil
}

func (us *UserRepository) Delete(ctx context.Context, userId int) error {
	_, err := us.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, userId)

	if err != nil {
		return fmt.Errorf("error deleting user: %v", err)
	}

	return nil
}
