package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jameswhoughton/meals/internal/account"
	"github.com/jameswhoughton/meals/internal/meals"
	"github.com/jameswhoughton/meals/memory"
	"github.com/jameswhoughton/meals/web"
)

func TestUserReceives404WhenTryingToAssignAMealThatDoesNotExist(t *testing.T) {
	plannerRepository := memory.NewPlannerRepository()

	mealRepository := memory.MealRepository{}

	handler := web.PostEditDayHandler(newTestLogger(), plannerRepository, &mealRepository)

	user := account.User{}

	ctx := context.WithValue(context.Background(), "user", user)

	form := url.Values{}
	form.Add("meal_id", "1")

	postData := strings.NewReader(form.Encode())

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/planner/2025-01-01", postData)

	req.Header.Set("Content-type", "application/x-www-form-urlencoded")

	req.SetPathValue("date", "2025-01-01")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code: %d, got %d", http.StatusNotFound, result.StatusCode)
	}
}

func TestUserReceieves403IfTryingToAssignAMealTheyDoNotOwn(t *testing.T) {
	plannerRepository := memory.NewPlannerRepository()

	mealRepository := memory.MealRepository{
		Store: []meals.Meal{
			{
				Id:     1,
				UserId: 2,
			},
		},
	}

	handler := web.PostEditDayHandler(newTestLogger(), plannerRepository, &mealRepository)

	user := account.User{}

	ctx := context.WithValue(context.Background(), "user", user)

	form := url.Values{}
	form.Add("meal_id", "1")

	postData := strings.NewReader(form.Encode())

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/planner/2025-01-01", postData)

	req.Header.Set("Content-type", "application/x-www-form-urlencoded")

	req.SetPathValue("date", "2025-01-01")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status code: %d, got %d", http.StatusForbidden, result.StatusCode)
	}
}
