package auth_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jameswhoughton/meals/internal/auth"
	"github.com/jameswhoughton/meals/memory"
)

func strPtr(v string) *string { return &v }

func TestCreateUserFailsIfFormIsInvalid(t *testing.T) {

	type testCase struct {
		description    string
		form           auth.UserFormCreate
		expectedErrors []string
	}

	cases := []testCase{
		{
			description: "Mismatched password",
			form: auth.UserFormCreate{
				Name:            "John Smith",
				Password:        "password",
				PasswordConfirm: "pssword",
				Email:           "john@example.com",
			},
			expectedErrors: []string{"Password"},
		},
		{
			description: "Fields missing",
			form: auth.UserFormCreate{
				Name:            "",
				Password:        "",
				PasswordConfirm: "",
				Email:           "",
			},
			expectedErrors: []string{"Name", "Email", "Password"},
		},
		{
			description: "Fields too long",
			form: auth.UserFormCreate{
				Name:            strings.Repeat("A", 256),
				Password:        strings.Repeat("A", 256),
				PasswordConfirm: strings.Repeat("A", 256),
				Email:           strings.Repeat("A", 256),
			},
			expectedErrors: []string{"Name", "Email", "Password"},
		},
		{
			description: "Email in use",
			form: auth.UserFormCreate{
				Name:            "Paul",
				Password:        "aaabbbcccd",
				PasswordConfirm: "aaabbbcccd",
				Email:           "paul@example.com",
			},
			expectedErrors: []string{"Email"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.description, func(t *testing.T) {
			userRepo := memory.UserRepository{
				Store: []auth.User{
					{
						Id:    1,
						Email: "paul@example.com",
					},
				},
			}
			service := auth.NewUserService(&userRepo, &memory.SessionRepository{}, 0)

			_, err := service.CreateUser(&testCase.form)

			if err == nil {
				t.Error("Expected error, got none")
			}

			if !errors.Is(err, auth.ErrorFormInvalid{}) {
				t.Errorf("Expected validation error, got: %v", err)
			}

			for _, expectedError := range testCase.expectedErrors {
				if testCase.form.Errors[expectedError] == "" {
					t.Errorf("Expected validation error for '%s' missing", expectedError)
				}
			}
		})
	}
}

func TestUpdateUserFailsIfFormIsInvalid(t *testing.T) {

	type testCase struct {
		description    string
		form           auth.UserFormUpdate
		expectedErrors []string
	}

	strPtr := func(v string) *string { return &v }

	cases := []testCase{
		{
			description: "Mismatched password",
			form: auth.UserFormUpdate{
				Id:              1,
				Name:            "John Smith",
				Password:        strPtr("password"),
				PasswordConfirm: "pssword",
				Email:           "john@example.com",
			},
			expectedErrors: []string{"Password"},
		},
		{
			description: "Fields too long",
			form: auth.UserFormUpdate{
				Id:              1,
				Name:            strings.Repeat("A", 256),
				Password:        strPtr(strings.Repeat("A", 256)),
				PasswordConfirm: strings.Repeat("A", 256),
				Email:           strings.Repeat("A", 256),
			},
			expectedErrors: []string{"Name", "Email", "Password"},
		},
		{
			description: "Email in use",
			form: auth.UserFormUpdate{
				Id:    1,
				Email: "paul@example.com",
			},
			expectedErrors: []string{"Email"},
		},
		{
			description: "Meal start day invalid",
			form: auth.UserFormUpdate{
				Name:            "Paul",
				Password:        strPtr("aaabbbcccd"),
				PasswordConfirm: "aaabbbcccd",
				Email:           "paul123@example.com",
				MealStartDay:    10,
			},
			expectedErrors: []string{"MealStartDay"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.description, func(t *testing.T) {
			userRepo := memory.UserRepository{
				Store: []auth.User{
					{
						Id: 1,
					},
					{
						Id:    2,
						Email: "paul@example.com",
					},
				},
			}

			service := auth.NewUserService(&userRepo, &memory.SessionRepository{}, 0)

			err := service.UpdateUser(&testCase.form)

			if err == nil {
				t.Error("Expected error, got none")
			}

			if !errors.Is(err, auth.ErrorFormInvalid{}) {
				t.Errorf("Expected validation error, got: %v", err)
			}

			for _, expectedError := range testCase.expectedErrors {
				if testCase.form.Errors[expectedError] == "" {
					t.Errorf("Expected validation error for '%s' missing", expectedError)
				}
			}
		})
	}
}

// test can create, get and update a user
func TestCanCreateGetAndUpdateAUser(t *testing.T) {
	service := auth.NewUserService(&memory.UserRepository{}, &memory.SessionRepository{}, 0)

	form := auth.UserFormCreate{
		Name:            "John Smith",
		Email:           "john.smith@example.com",
		Password:        "password123",
		PasswordConfirm: "password123",
	}

	user, err := service.CreateUser(&form)

	if err != nil {
		t.Errorf("Unexpected error creating user: %v", err)
	}

	if form.Name != user.Name {
		t.Errorf("Expected name: %s, got %s", form.Name, user.Name)
	}

	if form.Email != user.Email {
		t.Errorf("Expected email: %s, got %s", form.Email, user.Email)
	}

	if form.Password == user.Password {
		t.Errorf("Password not hashed %s - %s", form.Password, user.Password)
	}

	user, err = service.GetUserFromCredentials(form.Email, form.Password)

	if err != nil {
		t.Errorf("Unexpected error fetching user: %v", err)
	}

	if form.Name != user.Name {
		t.Errorf("Expected name: %s, got %s", form.Name, user.Name)
	}

	if form.Email != user.Email {
		t.Errorf("Expected email: %s, got %s", form.Email, user.Email)
	}

	if form.Password == user.Password {
		t.Errorf("Password not hashed %s - %s", form.Password, user.Password)
	}

	updateForm := auth.UserFormUpdate{
		Id:              user.Id,
		Name:            "Steve Smith",
		Email:           "steve.smith@example.com",
		MealStartDay:    4,
		Password:        strPtr("PASSWORD456!"),
		PasswordConfirm: "PASSWORD456!",
	}

	err = service.UpdateUser(&updateForm)

	if err != nil {
		t.Errorf("Unexpected error updating user: %v", err)
	}

	user, err = service.GetUserFromCredentials(updateForm.Email, *updateForm.Password)

	if err != nil {
		t.Errorf("Unexpected error updating user: %v", err)
	}

	if updateForm.Name != user.Name {
		t.Errorf("Expected name: %s, got %s", updateForm.Name, user.Name)
	}

	if updateForm.Email != user.Email {
		t.Errorf("Expected email: %s, got %s", updateForm.Email, user.Email)
	}

	if updateForm.MealStartDay != user.MealStartDay {
		t.Errorf("Expected start day: %d, got %d", updateForm.MealStartDay, user.MealStartDay)
	}

	if *updateForm.Password == user.Password {
		t.Errorf("Password not hashed %s - %s", *updateForm.Password, user.Password)
	}
}

// test returns expected error if a user does not exist
func TestReturnExpectedErrorIfUserDoesNotExist(t *testing.T) {
	service := auth.NewUserService(&memory.UserRepository{}, &memory.SessionRepository{}, 0)

	_, err := service.GetUserFromCredentials("", "")

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if !errors.Is(err, auth.ErrorUserNotFound{}) {
		t.Errorf("Expected ErrorUserNotFound, got %v", err)
	}
}

// Test any expired sessions are removed when a new session is created
func TestAnyExpiredSessionsAreRemovedWhenANewSessionIsCreated(t *testing.T) {
	sessionLifetime := 3600
	userRepo := memory.UserRepository{
		Store: []auth.User{
			{
				Id: 1,
			},
		},
	}

	sessionRepo := memory.SessionRepository{
		Store: []auth.Session{
			// Expired session
			{
				UserId:    2,
				CreatedAt: time.Now().Add(-time.Duration((sessionLifetime + 1) * 1000)),
			},
		},
	}

	service := auth.NewUserService(&userRepo, &sessionRepo, sessionLifetime)

	newSession, err := service.CreateSession(1)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(sessionRepo.Store) != 1 {
		t.Errorf("Expected only 1 session, found %d", len(sessionRepo.Store))
	}

	if newSession.SessionId != sessionRepo.Store[0].SessionId {
		t.Errorf("Expected session ID %s, found %s", newSession.SessionId, sessionRepo.Store[0].SessionId)
	}
}
