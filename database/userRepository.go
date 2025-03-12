package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jameswhoughton/meals/internal/auth"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRespository(db *sql.DB) *UserRepository {
	return &UserRepository{db}
}

func (us *UserRepository) Get(filter auth.UserGet) (auth.User, error) {
	var user auth.User
	var wheres []string

	err := filter.Validate()

	if err != nil {
		return auth.User{}, fmt.Errorf("UserRepository.Get invalid parameter: %w", err)
	}

	var values []any

	if filter.Id != nil {
		wheres = append(wheres, "id = ?")
		values = append(values, *filter.Id)
	}

	if filter.Email != nil {
		wheres = append(wheres, "email = ?")
		values = append(values, *filter.Email)
	}

	query := "SELECT id, name, email, password FROM users WHERE 1 = 1 AND " + strings.Join(wheres, " AND ")

	err = us.db.QueryRow(query, values...).Scan(&user.Id, &user.Name, &user.Email, &user.Password)

	if err != nil {
		if err == sql.ErrNoRows {
			return auth.User{}, auth.ErrorUserNotFound{}
		}

		return auth.User{}, fmt.Errorf("UserRepository.Get: error getting user: %v", err)
	}

	return user, nil
}

func (us *UserRepository) Create(user auth.User) (auth.User, error) {
	result, err := us.db.Exec(`
		INSERT INTO users 
		(email, name, password, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, user.Email, user.Name, user.Password, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		return auth.User{}, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return auth.User{}, err
	}

	user.Id = int(id)

	return user, nil
}

func (us *UserRepository) Update(id int, user auth.UserUpdate) error {
	_, err := us.Get(auth.UserGet{Id: &id})

	if err != nil {
		return err
	}

	updates := []string{
		"updated_at = ?",
	}

	values := []any{
		time.Now(),
	}

	if v := user.Email; v != nil {
		updates = append(updates, "email = ?")
		values = append(values, *v)
	}

	if v := user.Name; v != nil {
		updates = append(updates, "name = ?")
		values = append(values, *v)
	}

	if v := user.Password; v != nil {
		updates = append(updates, "password = ?")
		values = append(values, *v)
	}

	values = append(values, id)

	query := "UPDATE users SET " + strings.Join(updates, ",") + " WHERE id = ?"

	_, err = us.db.Exec(query, values...)

	if err != nil {
		return fmt.Errorf("error updating user: %v", err)
	}

	return nil
}
