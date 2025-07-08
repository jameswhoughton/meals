package meals

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	Id           int
	Name         string
	Email        string
	Password     string
	MealStartDay int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserUpdate struct {
	Id           int
	Name         string
	Email        string
	Password     *string
	MealStartDay int
	UpdatedAt    time.Time
}

type UserRepository interface {

	// Get a user that matches the given id
	//
	// Returns ErrUserNotFound if the id does not match any users
	GetById(ctx context.Context, id int) (User, error)

	// Get a user that matches the given email
	//
	// Returns ErrUserNotFound if the email does not match any users
	GetByEmail(ctx context.Context, email string) (User, error)

	// Create a new user
	//
	// Returns the new user if created successfully.
	Create(ctx context.Context, user User) (User, error)

	// Updates an existing user
	//
	// Returns ErrUserNotFound if the given user.Id does not exist.
	Update(ctx context.Context, user UserUpdate) error

	// Deletes an existing user
	//
	// Does nothing if the userId does not exist.
	Delete(ctx context.Context, userId int) error
}

func validateName(name string) (bool, string) {
	if len(name) == 0 {
		return false, "name cannot be empty"
	}

	if len(name) > 255 {
		return false, "name cannot contain more that 255 characters"
	}

	return true, ""
}

func validatePassword(password, passwordConfirm string) (bool, string) {
	if password != passwordConfirm {
		return false, "password and confirm do not match"
	}

	if len(password) < 10 {
		return false, "password is too short (must be more than 10 characters)"
	}

	if len(password) > 255 {
		return false, "password cannot contain more than 255 characters"
	}

	return true, ""
}

func validateEmail(email string) (bool, string) {
	if len(email) == 0 {
		return false, "email cannot be empty"
	}

	if len(email) > 255 {
		return false, "email cannot contain more that 255 characters"
	}

	return true, ""

}

func validateMealStartDay(day int) (bool, string) {
	if day < 0 || day > 6 {
		return false, "meal start day is not valid day of the week"
	}

	return true, ""
}

var ErrUserFormInvalid = errors.New("form has validation errors")

type UserFormUpdate struct {
	Id              int
	Password        *string `json:"-"`
	PasswordConfirm string  `json:"-"`
	Email           string
	Name            string
	MealStartDay    int
	Errors          map[string]string
}

func (f *UserFormUpdate) Validate(ctx context.Context, currentUser User, repo UserRepository) bool {
	f.Errors = make(map[string]string)

	if f.Password != nil {
		if passes, message := validatePassword(*f.Password, f.PasswordConfirm); !passes {
			f.Errors["Password"] = message
		}
	}

	if passes, message := validateName(f.Name); !passes {
		f.Errors["Name"] = message
	}

	if currentUser.Email != f.Email {
		if passes, message := validateEmail(f.Email); !passes {
			f.Errors["Email"] = message
		}

		existingUser, err := repo.GetByEmail(ctx, f.Email)

		if err == nil && existingUser.Id != currentUser.Id {
			f.Errors["Email"] = "email already in use"
		}
	}

	if passes, message := validateMealStartDay(f.MealStartDay); !passes {
		f.Errors["MealStartDay"] = message
	}

	return len(f.Errors) == 0
}

type UserFormCreate struct {
	Id              int
	Password        string `json:"-"`
	PasswordConfirm string `json:"-"`
	Email           string
	Name            string
	Errors          map[string]string
}

func (f *UserFormCreate) Validate(ctx context.Context, repo UserRepository) bool {
	f.Errors = make(map[string]string)

	if passes, message := validatePassword(f.Password, f.PasswordConfirm); !passes {
		f.Errors["Password"] = message
	}

	if passes, message := validateName(f.Name); !passes {
		f.Errors["Name"] = message
	}

	if passes, message := validateEmail(f.Email); !passes {
		f.Errors["Email"] = message
	}

	user, _ := repo.GetByEmail(ctx, f.Email)

	if user.Id > 0 {
		f.Errors["Email"] = "email already in use"
	}

	return len(f.Errors) == 0
}

func HashPassword(password string) []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return hash
}

func NewUserService(repo UserRepository) UserService {
	return UserService{
		account: repo,
	}
}

type UserService struct {
	account UserRepository
}

func (us *UserService) GetUserFromCredentials(ctx context.Context, email, password string) (User, error) {
	user, err := us.account.GetByEmail(ctx, email)

	if err != nil {
		return User{}, fmt.Errorf("failed to get user with email=%s: %w", email, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	if err != nil {
		return User{}, fmt.Errorf("credentials invalid: %v", err)
	}

	return user, nil
}

func (us *UserService) CreateUser(ctx context.Context, form *UserFormCreate) (User, error) {
	if !form.Validate(ctx, us.account) {
		return User{}, ErrUserFormInvalid
	}

	var user User

	user.Name = form.Name
	user.Email = form.Email
	user.Password = string(HashPassword(form.Password))
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	return us.account.Create(ctx, user)
}

func (us *UserService) UpdateUser(ctx context.Context, form *UserFormUpdate) error {
	currentUser, err := us.account.GetById(ctx, form.Id)

	if err != nil {
		return fmt.Errorf("failed to get user with id=%d: %w", form.Id, err)
	}

	if !form.Validate(ctx, currentUser, us.account) {
		return ErrUserFormInvalid
	}

	toUpdate := UserUpdate{
		Id:           form.Id,
		Name:         form.Name,
		Email:        form.Email,
		MealStartDay: form.MealStartDay,
	}

	if v := form.Password; v != nil {
		hashedPassword := string(HashPassword(*form.Password))

		toUpdate.Password = &hashedPassword
	}

	toUpdate.UpdatedAt = time.Now()

	return us.account.Update(ctx, toUpdate)
}
