package web_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/jameswhoughton/meals"
	"github.com/jameswhoughton/meals/memory"
	"github.com/jameswhoughton/meals/web"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestGetUserHandlerReturnsServerErrorIfUserMissingFromContext(t *testing.T) {
	templateFiles := fstest.MapFS{
		"templates/layout.gohtml":             {Data: []byte{}},
		"templates/navigation.gohtml":         {Data: []byte{}},
		"templates/pages/auth/account.gohtml": {Data: []byte{}},
	}

	sessionRepository := memory.SessionRepository{}
	userRepository := memory.UserRepository{}
	sessionService := web.NewSessionService(&userRepository, &sessionRepository, 3600)

	handler := web.GetAccountHandler(newTestLogger(), templateFiles, *sessionService)

	ctx := context.Background()

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/user", nil)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code: %d, got %d", http.StatusInternalServerError, result.StatusCode)
	}
}

func TestGetUserHandlerReturnsOkIfUserInContext(t *testing.T) {
	templateFiles := fstest.MapFS{
		"templates/layout.gohtml":          {Data: []byte{}},
		"templates/navigation.gohtml":      {Data: []byte{}},
		"templates/pages/auth/user.gohtml": {Data: []byte{}},
	}

	user := meals.User{
		Name:  "John Smith",
		Email: "john.smith@example.com",
	}

	sessionRepository := memory.SessionRepository{}
	userRepository := memory.UserRepository{}

	createdUser, err := userRepository.Create(context.Background(), user)

	if err != nil {
		t.Errorf("unexpected error creating user: %v", err)
	}

	sessionService := web.NewSessionService(&userRepository, &sessionRepository, 3600)

	handler := web.GetAccountHandler(newTestLogger(), templateFiles, *sessionService)

	ctx := context.WithValue(context.Background(), "user", createdUser)

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/user", nil)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusOK {
		t.Errorf("Expected status code: %d, got %d", http.StatusOK, result.StatusCode)
	}
}

func TestPostUserHandlerReturnsInternalServerErrorIfUserMissingFromContext(t *testing.T) {
	userRepository := memory.UserRepository{}
	sessionRepository := memory.SessionRepository{}
	sessionService := web.NewSessionService(&userRepository, &sessionRepository, 3600)
	userService := meals.NewUserService(&userRepository)
	logger := newTestLogger()

	handler := web.PutAccountHandler(logger, userService, *sessionService)

	ctx := context.Background()

	form := url.Values{}
	form.Add("name", "new name")
	form.Add("email", "new email")

	postData := strings.NewReader(form.Encode())

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/user", postData)

	req.Header.Set("Content-type", "application/x-www-form-urlencoded")

	req.ParseForm()

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code: %d, got %d", http.StatusInternalServerError, result.StatusCode)
	}
}

func TestPostUserHandlerReturnsOkAndUpdatesUser(t *testing.T) {
	userRepository := memory.UserRepository{}
	sessionRepository := memory.SessionRepository{}

	sessionService := web.NewSessionService(&userRepository, &sessionRepository, 3600)
	userService := meals.NewUserService(&userRepository)

	createdUser, err := userService.CreateUser(context.Background(), &meals.UserFormCreate{
		Name:            "John",
		Email:           "john@example.com",
		Password:        "password123",
		PasswordConfirm: "password123",
	})

	if err != nil {
		t.Errorf("unexpected error creating user")
	}

	ctx := context.WithValue(context.Background(), "user", createdUser)

	logger := newTestLogger()
	handler := web.PutAccountHandler(logger, userService, *sessionService)

	newEmail := "john.99@example.com"

	form := url.Values{}
	form.Add("name", createdUser.Name)
	form.Add("email", newEmail)
	form.Add("mealStartDay", strconv.Itoa(createdUser.MealStartDay))

	postData := strings.NewReader(form.Encode())

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/user", postData)

	req.Header.Set("Content-type", "application/x-www-form-urlencoded")

	req.ParseForm()

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusFound {
		t.Errorf("Expected status code: %d, got %d", http.StatusFound, result.StatusCode)
	}

	updatedUser, _ := userRepository.GetById(context.Background(), createdUser.Id)

	if updatedUser.Email != newEmail {
		t.Errorf("expected email %s, got %s", newEmail, updatedUser.Email)
	}
}
