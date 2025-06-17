package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jameswhoughton/meals/internal/account"
)

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRespository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db}
}

func (us *AccountRepository) GetById(ctx context.Context, id int) (account.User, error) {
	var user account.User

	err := us.db.QueryRowContext(
		ctx,
		`SELECT id, name, email, meal_start_day, password FROM users WHERE id = ?`,
		id).Scan(&user.Id, &user.Name, &user.Email, &user.MealStartDay, &user.Password)

	if err != nil {
		if err == sql.ErrNoRows {
			return account.User{}, account.ErrUserNotFound
		}

		return account.User{}, fmt.Errorf("failed to query user with id=%d: %w", id, err)
	}

	return user, nil
}

func (us *AccountRepository) GetByEmail(ctx context.Context, email string) (account.User, error) {
	var user account.User

	err := us.db.QueryRowContext(
		ctx,
		`SELECT id, name, email, meal_start_day, password FROM users WHERE email = ?`,
		email).Scan(&user.Id, &user.Name, &user.Email, &user.MealStartDay, &user.Password)

	if err != nil {
		if err == sql.ErrNoRows {
			return account.User{}, account.ErrUserNotFound
		}

		return account.User{}, fmt.Errorf("failed to query user with email=%s: %w", email, err)
	}

	return user, nil
}

func (us *AccountRepository) Create(ctx context.Context, user account.User) (account.User, error) {
	result, err := us.db.ExecContext(ctx, `
		INSERT INTO users 
		(email, name, password, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, user.Email, user.Name, user.Password, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		return account.User{}, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return account.User{}, err
	}

	user.Id = int(id)

	return user, nil
}

func (us *AccountRepository) Update(ctx context.Context, user account.UserUpdate) error {
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

func (us *AccountRepository) Delete(ctx context.Context, userId int) error {
	_, err := us.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, userId)

	if err != nil {
		return fmt.Errorf("error deleting user: %v", err)
	}

	return nil
}
