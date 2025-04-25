package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id           int
	Name         string
	Email        string
	Password     string
	MealStartDay int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func GenerateKey() string {
	key := make([]byte, 32)
	rand.Read(key)

	return base64.StdEncoding.EncodeToString(key)
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

type UserFormUpdate struct {
	Id              int
	Password        *string `json:"-"`
	PasswordConfirm string  `json:"-"`
	Email           string
	Name            string
	MealStartDay    int
	Errors          map[string]string
}

func (f *UserFormUpdate) Validate(currentUser User, userRepository UserRepository) bool {
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

		existingUser, err := userRepository.Get(UserGet{Email: &f.Email})

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

func (f *UserFormCreate) Validate(userRepository UserRepository) bool {
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

	user, _ := userRepository.Get(UserGet{Email: &f.Email})

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

func NewUserService(userRepo UserRepository, sessionRepo SessionRepository, sessionLifetime int) UserService {
	return UserService{
		userRepo:        userRepo,
		sessionRepo:     sessionRepo,
		sessionLifetime: sessionLifetime,
	}
}

type UserService struct {
	userRepo        UserRepository
	sessionRepo     SessionRepository
	sessionLifetime int
}

func (us *UserService) GetUserFromCredentials(email, password string) (User, error) {
	user, err := us.userRepo.Get(UserGet{Email: &email})

	if err != nil {
		return User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	if err != nil {
		return User{}, fmt.Errorf("Credentials invalid: %v", err)
	}

	return user, nil
}

func (us *UserService) CreateUser(form *UserFormCreate) (User, error) {
	if !form.Validate(us.userRepo) {
		return User{}, ErrorFormInvalid{}
	}

	var user User

	user.Name = form.Name
	user.Email = form.Email
	user.Password = string(HashPassword(form.Password))
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	createdUser, err := us.userRepo.Create(user)

	if err != nil {
		return User{}, fmt.Errorf("failed to create user: %v", err)
	}

	return createdUser, nil
}

type ErrorFormInvalid struct {
}

func (e ErrorFormInvalid) Error() string {
	return "The form contains errors"
}

func (us *UserService) UpdateUser(form *UserFormUpdate) error {
	currentUser, err := us.userRepo.Get(UserGet{Id: &form.Id})

	if err != nil {
		return fmt.Errorf("User update - error fetching current user model: %w", err)
	}

	if !form.Validate(currentUser, us.userRepo) {
		return ErrorFormInvalid{}
	}

	toUpdate := UserUpdate{}

	toUpdate.Name = form.Name
	toUpdate.Email = form.Email
	toUpdate.MealStartDay = form.MealStartDay

	if v := form.Password; v != nil {
		hashedPassword := string(HashPassword(*form.Password))

		toUpdate.Password = &hashedPassword
	}

	toUpdate.UpdatedAt = time.Now()

	return us.userRepo.Update(currentUser.Id, toUpdate)
}

func (us *UserService) sessionExpired() time.Time {
	return time.Now().Add(time.Second * time.Duration(-us.sessionLifetime))
}

func (us *UserService) CreateSession(userId int) (Session, error) {
	// Remove any sessions that have expired
	us.sessionRepo.DestroyExpired(us.sessionExpired())

	var session Session

	session.SessionId = GenerateKey()
	session.UserId = userId
	session.CreatedAt = time.Now()
	session.UpdatedAt = session.CreatedAt

	session, err := us.sessionRepo.Create(session)

	if err != nil {
		return Session{}, err
	}

	return session, nil
}

func (us *UserService) GetUserFromSession(sessionId string) (User, error) {
	session, err := us.sessionRepo.Get(sessionId, us.sessionExpired())

	if err != nil {
		return User{}, fmt.Errorf("error fetching session: %w", err)
	}

	user, err := us.userRepo.Get(UserGet{Id: &session.UserId})

	if err != nil {
		return User{}, fmt.Errorf("error fetching user: %w", err)
	}

	return user, nil
}
