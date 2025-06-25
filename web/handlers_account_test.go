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

	"github.com/jameswhoughton/meals/internal/account"
	"github.com/jameswhoughton/meals/memory"
	"github.com/jameswhoughton/meals/web"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestGetAccountHandlerReturnsServerErrorIfUserMissingFromContext(t *testing.T) {
	templateFiles := fstest.MapFS{
		"templates/layout.gohtml":             {Data: []byte{}},
		"templates/navigation.gohtml":         {Data: []byte{}},
		"templates/pages/auth/account.gohtml": {Data: []byte{}},
	}

	sessionRepository := memory.SessionRepository{}
	accountRepository := memory.AccountRepository{}
	sessionService := web.NewSessionService(&accountRepository, &sessionRepository, 3600)

	handler := web.GetAccountHandler(newTestLogger(), templateFiles, *sessionService)

	ctx := context.Background()

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/account", nil)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code: %d, got %d", http.StatusInternalServerError, result.StatusCode)
	}
}

func TestGetAccountHandlerReturnsOkIfUserInContext(t *testing.T) {
	templateFiles := fstest.MapFS{
		"templates/layout.gohtml":             {Data: []byte{}},
		"templates/navigation.gohtml":         {Data: []byte{}},
		"templates/pages/auth/account.gohtml": {Data: []byte{}},
	}

	user := account.User{
		Name:  "John Smith",
		Email: "john.smith@example.com",
	}

	sessionRepository := memory.SessionRepository{}
	accountRepository := memory.AccountRepository{}

	createdUser, err := accountRepository.Create(context.Background(), user)

	if err != nil {
		t.Errorf("unexpected error creating user: %v", err)
	}

	sessionService := web.NewSessionService(&accountRepository, &sessionRepository, 3600)

	handler := web.GetAccountHandler(newTestLogger(), templateFiles, *sessionService)

	ctx := context.WithValue(context.Background(), "user", createdUser)

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/account", nil)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusOK {
		t.Errorf("Expected status code: %d, got %d", http.StatusOK, result.StatusCode)
	}
}

func TestPostAccountHandlerReturnsInternalServerErrorIfUserMissingFromContext(t *testing.T) {
	accountRepository := memory.AccountRepository{}
	sessionRepository := memory.SessionRepository{}
	sessionService := web.NewSessionService(&accountRepository, &sessionRepository, 3600)
	accountService := account.NewService(&accountRepository)
	logger := newTestLogger()

	handler := web.PutAccountHandler(logger, accountService, *sessionService)

	ctx := context.Background()

	form := url.Values{}
	form.Add("name", "new name")
	form.Add("email", "new email")

	postData := strings.NewReader(form.Encode())

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/account", postData)

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

func TestPostAccountHandlerReturnsOkAndUpdatesAccount(t *testing.T) {
	accountRepository := memory.AccountRepository{}
	sessionRepository := memory.SessionRepository{}

	sessionService := web.NewSessionService(&accountRepository, &sessionRepository, 3600)
	accountService := account.NewService(&accountRepository)

	createdUser, err := accountService.CreateUser(context.Background(), &account.UserFormCreate{
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
	handler := web.PutAccountHandler(logger, accountService, *sessionService)

	newEmail := "john.99@example.com"

	form := url.Values{}
	form.Add("name", createdUser.Name)
	form.Add("email", newEmail)
	form.Add("mealStartDay", strconv.Itoa(createdUser.MealStartDay))

	postData := strings.NewReader(form.Encode())

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/account", postData)

	req.Header.Set("Content-type", "application/x-www-form-urlencoded")

	req.ParseForm()

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusFound {
		t.Errorf("Expected status code: %d, got %d", http.StatusFound, result.StatusCode)
	}

	updatedUser, _ := accountRepository.GetById(context.Background(), createdUser.Id)

	if updatedUser.Email != newEmail {
		t.Errorf("expected email %s, got %s", newEmail, updatedUser.Email)
	}
}
