package meals_test

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

func TestPostMealHandlerCraetesAMeal(t *testing.T) {
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
			form.Add("isMain", strconv.Itoa(i))
		}
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

	storedMeal, err := mealRepository.Get(1)

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
		sort := func(x, y meals.MealIngredient) int {
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
				Ingredients: []meals.MealIngredient{
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
	ingredientRepository := memory.IngredientRepository{}
	service := meals.NewService(&mealRepository, &ingredientRepository)

	handler := meals.PutMealHandler(service)

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
		Ingredients: []meals.MealIngredient{
			{
				Id:       1,
				Name:     "Chicken breast",
				Quantity: 2,
				IsMain:   true,
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

	for i, ingredient := range mealToUpdate.Ingredients {
		form.Add("ingredientId", strconv.Itoa(ingredient.Id))
		form.Add("ingredientName", ingredient.Name)
		form.Add("ingredientQuantity", strconv.Itoa(ingredient.Quantity))
		form.Add("ingredientUnit", ingredient.Unit)

		if ingredient.IsMain {
			form.Add("isMain", strconv.Itoa(i))
		}
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

	storedMeal, err := mealRepository.Get(14)

	if err != nil {
		t.Errorf("Unexpected error fetching updated meal: %v", err)
	}

	compareMeals(t, mealToUpdate, storedMeal)
}

func TestGetIngredientHandlerReturns404IfTheIngredientDoesNotExist(t *testing.T) {
	templateFiles := fstest.MapFS{
		"templates/layout.gohtml":                 {Data: []byte{}},
		"templates/navigation.gohtml":             {Data: []byte{}},
		"templates/pages/ingredients/edit.gohtml": {Data: []byte{}},
	}

	repository := memory.IngredientRepository{}

	handler := meals.GetIngredientHandler(templateFiles, &repository)

	ctx := context.WithValue(context.Background(), "userId", 1)

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/ingredient/1", nil)

	req.SetPathValue("id", "1")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code: %d, got %d", http.StatusNotFound, result.StatusCode)
	}
}

func TestGetIngredientHandlerReturns403IfIngredientDoesNotBelongToTheUser(t *testing.T) {
	templateFiles := fstest.MapFS{
		"templates/layout.gohtml":                 {Data: []byte{}},
		"templates/navigation.gohtml":             {Data: []byte{}},
		"templates/pages/ingredients/edit.gohtml": {Data: []byte{}},
	}

	repository := memory.IngredientRepository{
		Store: []meals.Ingredient{
			{
				Id:     1,
				UserId: 2,
			},
		},
	}

	handler := meals.GetIngredientHandler(templateFiles, &repository)

	ctx := context.WithValue(context.Background(), "userId", 1)

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/ingredients/1", nil)

	req.SetPathValue("id", "1")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status code: %d, got %d", http.StatusForbidden, result.StatusCode)
	}
}

func TestPutIngredientHandlerReturns404IfIngredientDoesNotExist(t *testing.T) {
	ingredientRepository := memory.IngredientRepository{}
	mealRepository := memory.MealRepository{}

	service := meals.NewService(&mealRepository, &ingredientRepository)

	handler := meals.PutIngredientHandler(service)

	ctx := context.WithValue(context.Background(), "userId", 1)

	form := url.Values{}
	form.Add("name", "updated name")

	postData := strings.NewReader(form.Encode())

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/ingredient/1", postData)

	req.SetPathValue("id", "1")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code: %d, got %d", http.StatusNotFound, result.StatusCode)
	}
}

func TestPutIngredientHandlerReturns403IfIngredientDoesNotBelongToUser(t *testing.T) {
	mealRepository := memory.MealRepository{}
	ingredientRepository := memory.IngredientRepository{
		Store: []meals.Ingredient{
			{
				Id:   12,
				Name: "Old name",
			},
		},
	}
	service := meals.NewService(&mealRepository, &ingredientRepository)

	handler := meals.PutIngredientHandler(service)

	ctx := context.WithValue(context.Background(), "userId", 1)

	form := url.Values{}
	form.Add("name", "updated name")

	postData := strings.NewReader(form.Encode())

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/ingredients/12", postData)

	req.Header.Set("Content-type", "application/x-www-form-urlencoded")

	req.SetPathValue("id", "12")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status code: %d, got %d", http.StatusForbidden, result.StatusCode)
	}
}

func TestPutIngredientHandlerWithCorrectFormUpdatesIngredient(t *testing.T) {
	mealRepository := memory.MealRepository{}
	ingredientRepository := memory.IngredientRepository{
		Store: []meals.Ingredient{
			{
				Id:     12,
				UserId: 2,
				Name:   "Old name",
			},
		},
	}
	service := meals.NewService(&mealRepository, &ingredientRepository)

	handler := meals.PutIngredientHandler(service)

	ctx := context.WithValue(context.Background(), "userId", 2)

	form := url.Values{}
	form.Add("name", "updated name")

	postData := strings.NewReader(form.Encode())

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/ingredients/12", postData)

	req.Header.Set("Content-type", "application/x-www-form-urlencoded")

	req.SetPathValue("id", "12")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusFound {
		t.Errorf("Expected status code: %d, got %d", http.StatusFound, result.StatusCode)
	}

	fetchedIngredient, err := ingredientRepository.GetById(12)

	if err != nil {
		t.Errorf("Unexpected error fetching ingredient: %v", err)
	}

	if fetchedIngredient.Name != form.Get("name") {
		t.Errorf("Expected name to be %s, found %s", form.Get("name"), fetchedIngredient.Name)
	}
}

// GetTagHandler returns 404 if the tag does not exist

// GetTagHandler returns 403 if the tag does belong to the user

// PutTagHandler returns 404 if tag does not exist

// PutTagHandler returns 403 if tag does not belong to the user

// PutTagHandler with the correct form, updates the correct tag
