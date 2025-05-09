package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/jameswhoughton/meals/internal/meals"
	"github.com/jameswhoughton/meals/memory"
	"github.com/jameswhoughton/meals/web"
)

func TestGetMealHandlerReturns404IfMealDoesNotExist(t *testing.T) {
	templateFiles := fstest.MapFS{
		"templates/layout.gohtml":                    {Data: []byte{}},
		"templates/navigation.gohtml":                {Data: []byte{}},
		"templates/pages/meals/create_update.gohtml": {Data: []byte{}},
	}

	repository := memory.MealRepository{}

	handler := web.GetMealHandler(templateFiles, &repository)

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

	handler := web.GetMealHandler(templateFiles, &repository)

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
	tagRepository := memory.TagRepository{}
	service := meals.NewService(&mealRepository, &tagRepository)

	handler := web.PutMealHandler(service, &mealRepository)

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
	tagRepository := memory.TagRepository{}
	service := meals.NewService(&mealRepository, &tagRepository)

	handler := web.PutMealHandler(service, &mealRepository)

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

func TestPostMealHandlerCraetesAMeal(t *testing.T) {
	mealRepository := memory.MealRepository{}
	tagRepository := memory.TagRepository{}
	service := meals.NewService(&mealRepository, &tagRepository)

	handler := web.PostMealHandler(service)

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
		Ingredients: []meals.Ingredient{
			{
				Id:       1,
				Name:     "Chicken breast",
				Quantity: 2,
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

	for _, ingredient := range mealToCreate.Ingredients {
		form.Add("ingredientId", strconv.Itoa(ingredient.Id))
		form.Add("ingredientName", ingredient.Name)
		form.Add("ingredientQuantity", strconv.Itoa(ingredient.Quantity))
		form.Add("ingredientUnit", ingredient.Unit)
	}

	postData := strings.NewReader(form.Encode())

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/meals/create", postData)

	req.Header.Set("Content-type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusFound {
		t.Errorf("Expected status code: %d, got %d", http.StatusFound, result.StatusCode)
	}

	storedMeal, err := mealRepository.Get(context.Background(), 1)

	if err != nil {
		t.Errorf("Unexpected error fetching created meal: %v", err)
	}

	compareMeals(t, mealToCreate, storedMeal)
}

func compareMeals(t *testing.T, a, b meals.Meal) {
	if a.Name != b.Name {
		t.Errorf("Names do not match (%s - %s)", a.Name, b.Name)
	}

	if a.Notes != b.Notes {
		t.Errorf("Notes do not match (%s - %s)", a.Notes, b.Notes)
	}

	if len(a.Ingredients) == len(b.Ingredients) {
		sort := func(x, y meals.Ingredient) int {
			if x.Id > y.Id {
				return 1
			}

			return -1
		}
		slices.SortFunc(a.Ingredients, sort)
		slices.SortFunc(b.Ingredients, sort)

		for i, aIngredient := range a.Ingredients {
			if aIngredient.Name != b.Ingredients[i].Name {
				t.Errorf("Ingredient names do not match (%s - %s)", aIngredient.Name, b.Ingredients[0].Name)
			}
		}
	} else {
		t.Errorf("Meals have a differing number of ingredients")
	}

	if len(a.Tags) == len(b.Tags) {
		sort := func(x, y meals.Tag) int {
			if x.Id > y.Id {
				return 1
			}

			return -1
		}
		slices.SortFunc(a.Tags, sort)
		slices.SortFunc(b.Tags, sort)

		for i, aTag := range a.Tags {
			if aTag.Name != b.Tags[i].Name {
				t.Errorf("Tag names do not match (%s - %s)", aTag.Name, b.Tags[0].Name)
			}
		}
	} else {
		t.Errorf("Meals have a differing number of tags")
	}
}

func TestPutMealHandlerWithCorrectFormUpdatesAMeal(t *testing.T) {
	mealRepository := memory.MealRepository{
		Store: []meals.Meal{
			{
				Id:     14,
				UserId: 1,
				Name:   "Old name",
				Notes:  "Old notes",
				Tags: []meals.Tag{
					{
						Id:   3,
						Name: "Old tag",
					},
				},
				Ingredients: []meals.Ingredient{
					{
						Id:       12,
						Name:     "Old ingredient",
						Quantity: 24,
						Unit:     "G",
					},
				},
			},
		},
	}
	tagRepository := memory.TagRepository{}
	service := meals.NewService(&mealRepository, &tagRepository)

	handler := web.PutMealHandler(service, &mealRepository)

	ctx := context.WithValue(context.Background(), "userId", 1)

	mealToUpdate := meals.Meal{
		Name:  "New name",
		Notes: "Yummy!",
		Tags: []meals.Tag{
			{
				Id:   1,
				Name: "Easy",
			},
		},
		Ingredients: []meals.Ingredient{
			{
				Id:       1,
				Name:     "Chicken breast",
				Quantity: 2,
			},
		},
	}

	form := url.Values{}

	form.Add("name", mealToUpdate.Name)
	form.Add("notes", mealToUpdate.Notes)

	for _, tag := range mealToUpdate.Tags {
		form.Add("tagId", strconv.Itoa(tag.Id))
		form.Add("tagName", tag.Name)
	}

	for _, ingredient := range mealToUpdate.Ingredients {
		form.Add("ingredientId", strconv.Itoa(ingredient.Id))
		form.Add("ingredientName", ingredient.Name)
		form.Add("ingredientQuantity", strconv.Itoa(ingredient.Quantity))
		form.Add("ingredientUnit", ingredient.Unit)
	}

	postData := strings.NewReader(form.Encode())

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/meals/14", postData)

	req.SetPathValue("id", "14")

	req.Header.Set("Content-type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusFound {
		t.Errorf("Expected status code: %d, got %d", http.StatusFound, result.StatusCode)
	}

	storedMeal, err := mealRepository.Get(context.Background(), 14)

	if err != nil {
		t.Errorf("Unexpected error fetching updated meal: %v", err)
	}

	compareMeals(t, mealToUpdate, storedMeal)
}

func TestPostMealDeleteReturns404IfMealDoesNotExist(t *testing.T) {
	mealRepository := memory.MealRepository{}

	handler := web.PostDeleteMealHandler(&mealRepository)

	ctx := context.WithValue(context.Background(), "userId", 1)

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/meals/1/delete", nil)

	req.SetPathValue("id", "1")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code: %d, got %d", http.StatusNotFound, result.StatusCode)
	}
}

func TestPostMealDeleteReturns403IfTheUserDoesNotOwnTheMeal(t *testing.T) {
	repository := memory.MealRepository{
		Store: []meals.Meal{
			{
				Id:     1,
				UserId: 2,
			},
		},
	}

	handler := web.PostDeleteMealHandler(&repository)

	ctx := context.WithValue(context.Background(), "userId", 1)

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/meals/1/delete", nil)

	req.SetPathValue("id", "1")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status code: %d, got %d", http.StatusForbidden, result.StatusCode)
	}
}

func TestPostMealDeleteDeletesMeal(t *testing.T) {
	repository := memory.MealRepository{
		Store: []meals.Meal{
			{
				Id:     1,
				UserId: 1,
			},
		},
	}

	handler := web.PostDeleteMealHandler(&repository)

	ctx := context.WithValue(context.Background(), "userId", 1)

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/meals/1/delete", nil)

	req.SetPathValue("id", "1")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusFound {
		t.Errorf("Expected status code: %d, got %d", http.StatusFound, result.StatusCode)
	}

	if len(repository.Store) != 0 {
		t.Errorf("Expected the store to be empty, found %d meals", len(repository.Store))
	}
}

// GetTagHandler returns 404 if the tag does not exist

// GetTagHandler returns 403 if the tag does belong to the user

// PutTagHandler returns 404 if tag does not exist

// PutTagHandler returns 403 if tag does not belong to the user

// PutTagHandler with the correct form, updates the correct tag
