package database

import (
	"database/sql"
	"fmt"

	"github.com/jameswhoughton/meals"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) UserService {
	return UserService{db}
}

func (us *UserService) Get(id int) (meals.User, error) {
	var user meals.User

	if err := us.db.QueryRow("SELECT * FROM users WHERE id = ?", id).Scan(&user); err != nil {
		if err == sql.ErrNoRows {
			return meals.User{}, fmt.Errorf("user %d not found", id)
		}

		return meals.User{}, fmt.Errorf("error fetching user %d: %v", id, err)
	}

	return user, nil
}

func (us *UserService) GetFromSessionId(sessionId string) (meals.User, error) {
	var user meals.User

	if err := us.db.QueryRow("SELECT u.id, u.email, u.password FROM sessions s LEFT JOIN users u ON s.user_id = u.id WHERE session_id = ?", sessionId).Scan(&user.Id, &user.Email, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			return meals.User{}, fmt.Errorf("session ID invalid")
		}

		return meals.User{}, fmt.Errorf("error fetching user: %v", err)
	}

	return user, nil
}

func (us *UserService) GetFromCredentials(email, password string) (meals.User, error) {
	user, err := us.GetFromEmail(email)

	if err != nil {
		return user, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	if err != nil {
		return meals.User{}, fmt.Errorf("credentials invalid")
	}

	return user, nil
}

func (us *UserService) GetFromEmail(email string) (meals.User, error) {
	var user meals.User

	if err := us.db.QueryRow("SELECT id, email, password FROM users WHERE email = ?", email).Scan(&user.Id, &user.Email, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			return meals.User{}, fmt.Errorf("credentials invalid")
		}

		return meals.User{}, fmt.Errorf("error fetching user %s: %v", email, err)
	}

	return user, nil
}

func (us *UserService) Add(user meals.User) (meals.User, error) {
	result, err := us.db.Exec("INSERT INTO users (email, password) VALUES (?, ?)", user.Email, user.Password)

	if err != nil {
		return meals.User{}, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return meals.User{}, err
	}

	return meals.User{
		Id:    int(id),
		Email: user.Email,
	}, nil
}

func (us *UserService) UpdateEmail(user meals.User, email string) error {
	if email == user.Email {
		return nil
	}

	_, err := us.db.Exec("UPDATE users SET email = ? WHERE id = ?", email, user.Id)

	if err != nil {
		return err
	}

	return nil
}

func (us *UserService) UpdatePassword(user meals.User, hash string) error {
	if hash == user.Password {
		return nil
	}

	_, err := us.db.Exec("UPDATE users SET password = ? WHERE id = ?", hash, user.Id)

	if err != nil {
		return err
	}

	return nil
}
