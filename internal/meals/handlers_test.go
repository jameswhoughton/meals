package meals_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/jameswhoughton/meals/internal/meals"
	"github.com/jameswhoughton/meals/memory"
)

func TestGetMealHandlerReturns404IfMealDoesNotExist(t *testing.T) {
	templateFiles := fstest.MapFS{
		"templates/layout.gohtml":                    {Data: []byte{}},
		"templates/navigation.gohtml":                {Data: []byte{}},
		"templates/pages/meals/create_update.gohtml": {Data: []byte{}},
	}

	repository := memory.MealRepository{}

	handler := meals.GetMealHandler(templateFiles, &repository)

	ctx := context.WithValue(context.Background(), "userId", 1)

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/meals/1", nil)

	req.SetPathValue("id", "1")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code: %d, got %d", http.StatusNotFound, result.StatusCode)
	}
}

func TestGetMealHandlerReturns403IfMealDoesNotBelongToUser(t *testing.T) {
	templateFiles := fstest.MapFS{
		"templates/layout.gohtml":                    {Data: []byte{}},
		"templates/navigation.gohtml":                {Data: []byte{}},
		"templates/pages/meals/create_update.gohtml": {Data: []byte{}},
	}

	repository := memory.MealRepository{
		Store: []meals.Meal{
			{
				Id:     1,
				UserId: 2,
			},
		},
	}

	handler := meals.GetMealHandler(templateFiles, &repository)

	ctx := context.WithValue(context.Background(), "userId", 1)

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/meals/", nil)

	req.SetPathValue("id", "1")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status code: %d, got %d", http.StatusForbidden, result.StatusCode)
	}

}

func TestPutMealHandlerReturns404IfMealDoesNotExist(t *testing.T) {
	mealRepository := memory.MealRepository{}
	ingredientRepository := memory.IngredientRepository{}
	service := meals.NewService(&mealRepository, &ingredientRepository)

	handler := meals.PutMealHandler(service)

	ctx := context.WithValue(context.Background(), "userId", 1)

	form := url.Values{}
	form.Add("name", "updated name")
	form.Add("ingredientId", "0")
	form.Add("ingredientName", "Carrots")
	form.Add("ingredientQuantity", "2")
	form.Add("ingredientUnit", "")
	form.Add("isMain", "0")

	postData := strings.NewReader(form.Encode())

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/meals/1", postData)

	req.Header.Set("Content-type", "application/x-www-form-urlencoded")

	req.ParseForm()

	fmt.Println(req.Form)

	req.SetPathValue("id", "1")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code: %d, got %d", http.StatusNotFound, result.StatusCode)
	}
}

func TestPutMealHandlerReturns403IfMealDoesNotBelongToTheUser(t *testing.T) {
	mealRepository := memory.MealRepository{
		Store: []meals.Meal{
			{
				Id:     20,
				UserId: 2,
			},
		},
	}
	ingredientRepository := memory.IngredientRepository{}
	service := meals.NewService(&mealRepository, &ingredientRepository)

	handler := meals.PutMealHandler(service)

	ctx := context.WithValue(context.Background(), "userId", 1)

	form := url.Values{}
	form.Add("name", "updated name")
	form.Add("ingredientId", "0")
	form.Add("ingredientName", "Carrots")
	form.Add("ingredientQuantity", "2")
	form.Add("ingredientUnit", "")
	form.Add("isMain", "0")

	postData := strings.NewReader(form.Encode())

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/meals/20", postData)

	req.Header.Set("Content-type", "application/x-www-form-urlencoded")

	req.SetPathValue("id", "20")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status code: %d, got %d", http.StatusForbidden, result.StatusCode)
	}
}

func PostMealHandlerCraetesAMeal(t *testing.T) {
	mealRepository := memory.MealRepository{}
	ingredientRepository := memory.IngredientRepository{}
	service := meals.NewService(&mealRepository, &ingredientRepository)

	handler := meals.PostMealHandler(service)

	ctx := context.WithValue(context.Background(), "userId", 1)

	mealToCreate := meals.Meal{
		Name:  "Chicken Pie",
		Notes: "Yummy!",
		Tags: []meals.Tag{
			{
				Id:   1,
				Name: "Easy",
			},
			{
				Id:   3,
				Name: "Family",
			},
		},
		Ingredients: []meals.MealIngredient{
			{
				Id:       1,
				Name:     "Chicken breast",
				Quantity: 2,
				IsMain:   true,
			},
			{
				Id:       2,
				Name:     "Leek",
				Quantity: 1,
			},
			{
				Id:       3,
				Name:     "Pastry",
				Quantity: 200,
				Unit:     "g",
			},
		},
	}

	form := url.Values{}

	form.Add("name", mealToCreate.Name)
	form.Add("notes", mealToCreate.Notes)

	for _, tag := range mealToCreate.Tags {
		form.Add("tagId", strconv.Itoa(tag.Id))
		form.Add("tagName", tag.Name)
	}

	for i, ingredient := range mealToCreate.Ingredients {
		form.Add("ingredientId", strconv.Itoa(ingredient.Id))
		form.Add("ingredientName", ingredient.Name)
		form.Add("ingredientQuantity", strconv.Itoa(ingredient.Quantity))
		form.Add("ingredientUnit", ingredient.Unit)

		if ingredient.IsMain {
			form.Add("mainIngredient", strconv.Itoa(i))
		}
	}

	postData := strings.NewReader(form.Encode())

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/meals/create", postData)

	req.Header.Set("Content-type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusOK {
		t.Errorf("Expected status code: %d, got %d", http.StatusOK, result.StatusCode)
	}

	// TODO check the meal exists correctly in the store
}

// PutMealHandler with the correct form updates the correct meal

// GetIngredientHandler returns 404 if the ingredient does not exist

// GetIngredientHandler returns 403 if the ingredient does belong to the user

// PutIngredientHandler returns 404 if ingredient does not exist

// PutIngredientHandler returns 403 if ingredient does not belong to the user

// PutIngredientHandler with the correct form, updates the correct ingredient

// GetTagHandler returns 404 if the tag does not exist

// GetTagHandler returns 403 if the tag does belong to the user

// PutTagHandler returns 404 if tag does not exist

// PutTagHandler returns 403 if tag does not belong to the user

// PutTagHandler with the correct form, updates the correct tag
