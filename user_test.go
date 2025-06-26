package meals_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jameswhoughton/meals"
	"github.com/jameswhoughton/meals/memory"
)

func strPtr(v string) *string { return &v }

func TestCreateUserFailsIfFormIsInvalid(t *testing.T) {

	type testCase struct {
		description    string
		form           meals.UserFormCreate
		expectedErrors []string
	}

	ctx := context.Background()

	cases := []testCase{
		{
			description: "Mismatched password",
			form: meals.UserFormCreate{
				Name:            "John Smith",
				Password:        "password",
				PasswordConfirm: "pssword",
				Email:           "john@example.com",
			},
			expectedErrors: []string{"Password"},
		},
		{
			description: "Fields missing",
			form: meals.UserFormCreate{
				Name:            "",
				Password:        "",
				PasswordConfirm: "",
				Email:           "",
			},
			expectedErrors: []string{"Name", "Email", "Password"},
		},
		{
			description: "Fields too long",
			form: meals.UserFormCreate{
				Name:            strings.Repeat("A", 256),
				Password:        strings.Repeat("A", 256),
				PasswordConfirm: strings.Repeat("A", 256),
				Email:           strings.Repeat("A", 256),
			},
			expectedErrors: []string{"Name", "Email", "Password"},
		},
		{
			description: "Email in use",
			form: meals.UserFormCreate{
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
			userRepo := memory.UserRepository{}

			userRepo.Create(ctx, meals.User{
				Id:    1,
				Email: "paul@example.com",
			})
			service := meals.NewUserService(&userRepo)

			_, err := service.CreateUser(ctx, &testCase.form)

			if err == nil {
				t.Error("Expected error, got none")
			}

			if !errors.Is(err, meals.ErrUserFormInvalid) {
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
		form           meals.UserFormUpdate
		expectedErrors []string
	}

	strPtr := func(v string) *string { return &v }

	ctx := context.Background()

	cases := []testCase{
		{
			description: "Mismatched password",
			form: meals.UserFormUpdate{
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
			form: meals.UserFormUpdate{
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
			form: meals.UserFormUpdate{
				Id:    1,
				Email: "paul@example.com",
			},
			expectedErrors: []string{"Email"},
		},
		{
			description: "Meal start day invalid",
			form: meals.UserFormUpdate{
				Id:              1,
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
			userRepo := memory.UserRepository{}

			userRepo.Create(ctx, meals.User{Id: 1})
			userRepo.Create(ctx, meals.User{
				Id:    2,
				Email: "paul@example.com",
			})

			service := meals.NewUserService(&userRepo)

			err := service.UpdateUser(ctx, &testCase.form)

			if err == nil {
				t.Error("Expected error, got none")
			}

			if !errors.Is(err, meals.ErrUserFormInvalid) {
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
	service := meals.NewUserService(&memory.UserRepository{})

	ctx := context.Background()

	form := meals.UserFormCreate{
		Name:            "John Smith",
		Email:           "john.smith@example.com",
		Password:        "password123",
		PasswordConfirm: "password123",
	}

	user, err := service.CreateUser(ctx, &form)

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

	user, err = service.GetUserFromCredentials(ctx, form.Email, form.Password)

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

	updateForm := meals.UserFormUpdate{
		Id:              user.Id,
		Name:            "Steve Smith",
		Email:           "steve.smith@example.com",
		MealStartDay:    4,
		Password:        strPtr("PASSWORD456!"),
		PasswordConfirm: "PASSWORD456!",
	}

	err = service.UpdateUser(ctx, &updateForm)

	if err != nil {
		t.Errorf("Unexpected error updating user: %v", err)
	}

	user, err = service.GetUserFromCredentials(ctx, updateForm.Email, *updateForm.Password)

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
	service := meals.NewUserService(&memory.UserRepository{})

	ctx := context.Background()

	_, err := service.GetUserFromCredentials(ctx, "", "")

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if !errors.Is(err, meals.ErrUserNotFound) {
		t.Errorf("Expected ErrorUserNotFound, got %v", err)
	}
}
