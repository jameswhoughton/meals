package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/jameswhoughton/meals"
	"github.com/jameswhoughton/meals/memory"
	"github.com/jameswhoughton/meals/web"
)

func TestUserReceives404WhenTryingToAssignAMealThatDoesNotExist(t *testing.T) {
	plannerRepository := memory.NewPlannerRepository()

	mealRepository := memory.MealRepository{}

	handler := web.PostEditDayHandler(newTestLogger(), plannerRepository, &mealRepository)

	user := meals.User{}

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

	user := meals.User{}

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

func TestCanAssignMealsToADate(t *testing.T) {
	plannerRepository := memory.NewPlannerRepository()

	mealRepository := memory.MealRepository{
		Store: []meals.Meal{
			{
				Id:     1,
				UserId: 2,
			},
			{
				Id:     2,
				UserId: 2,
			},
		},
	}

	plannerRepository.Meals = mealRepository.Store

	handler := web.PostEditDayHandler(newTestLogger(), plannerRepository, &mealRepository)

	user := meals.User{
		Id: 2,
	}

	ctx := context.WithValue(context.Background(), "user", user)

	form := url.Values{}
	form.Add("meal_id", "1")
	form.Add("meal_id", "2")

	postData := strings.NewReader(form.Encode())

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/planner/2025-01-01", postData)

	req.Header.Set("Content-type", "application/x-www-form-urlencoded")

	req.SetPathValue("date", "2025-01-01")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusFound {
		t.Errorf("expected status code: %d, got %d", http.StatusFound, result.StatusCode)
	}

	meals, err := plannerRepository.GetMealIdsInRange(ctx, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), user.Id)

	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(meals["2025-01-01"]) != 2 {
		t.Errorf("expected 2 meals assigned to date, got %d", len(meals))
	}
}
