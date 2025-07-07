package web

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"strconv"
	"text/template"

	"github.com/jameswhoughton/meals"
)

// Render the meals list page.
//
// Only the authenticated user's meals are visible
// Results can be filtered
func GetMealsHandler(logger *slog.Logger, templateFiles fs.FS, repo meals.MealRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(
			templateFiles,
			"templates/layout.gohtml",
			"templates/navigation.gohtml",
			"templates/pages/meals/list.gohtml",
		)

		if err != nil {
			w.Write([]byte("Template error: " + err.Error()))

			return
		}

		user := UserFromContext(r.Context())

		if user == nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"user missing from context",
			)

			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		queryString := r.URL.Query().Get("query")

		filter := meals.MealFilter{
			UserId: user.Id,
			Name:   &queryString,
		}

		results, err := repo.Find(r.Context(), filter)

		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to get meals",
				slog.Any("err", err),
				slog.Int("userId", user.Id),
				slog.Any("filters", filter),
			)

			http.Error(w, "server error", http.StatusInternalServerError)
		}

		success, err := getMessage(w, r, "success")

		type templateData struct {
			Title       string
			Success     string
			SearchQuery string
			Meals       []meals.Meal
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title:       "Meals",
			Success:     success,
			SearchQuery: queryString,
			Meals:       results,
		})
	})
}

// Render the edit meal page
//
// Only the owner of a meal can access this page.
// If a meal does not exist a 404 is returned.
// Expects the meal Id as a url path value with the name 'id'.
func GetMealEditHandler(logger *slog.Logger, templateFiles fs.FS, repo meals.MealRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(
			templateFiles,
			"templates/layout.gohtml",
			"templates/navigation.gohtml",
			"templates/pages/meals/create_update.gohtml",
		)

		if err != nil {
			w.Write([]byte("Template error: " + err.Error()))

			return
		}

		user := UserFromContext(r.Context())

		if user == nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"user missing from context",
			)

			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}
		mealId, _ := strconv.Atoi(r.PathValue("id"))

		meal, err := repo.Get(r.Context(), mealId)

		if err != nil {
			if errors.Is(err, meals.ErrMealNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)

				return
			}
		}

		if meal.UserId != user.Id {
			http.Error(w, "You do not have permission to access this page", http.StatusForbidden)

			return
		}

		formJson, err := getMessage(w, r, "formData")

		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"unable to fetch form data from cookie",
				slog.Any("err", err),
			)

			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		formData := meal

		if formJson != "" {
			json.Unmarshal([]byte(formJson), &formData)
		}

		type templateData struct {
			Title  string
			Form   meals.Meal
			Action string
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title:  meal.Name,
			Form:   formData,
			Action: "/meals/" + r.PathValue("id"),
		})
	})
}

// Render the view meal page
//
// Only the owner of a meal can access this page.
// If a meal does not exist a 404 is returned.
// Expects the meal Id as a url path value with the name 'id'.
func GetMealHandler(logger *slog.Logger, templateFiles fs.FS, repo meals.MealRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(
			templateFiles,
			"templates/layout.gohtml",
			"templates/navigation.gohtml",
			"templates/pages/meals/view.gohtml",
		)

		if err != nil {
			w.Write([]byte("Template error: " + err.Error()))

			return
		}

		user := UserFromContext(r.Context())

		if user == nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"user missing from context",
			)

			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}
		mealId, _ := strconv.Atoi(r.PathValue("id"))

		meal, err := repo.Get(r.Context(), mealId)

		if err != nil {
			if errors.Is(err, meals.ErrMealNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)

				return
			}
		}

		if meal.UserId != user.Id {
			http.Error(w, "You do not have permission to access this page", http.StatusForbidden)

			return
		}

		type templateData struct {
			Title string
			Meal  meals.Meal
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title: meal.Name,
			Meal:  meal,
		})
	})
}

// Render the create a meal page
func GetCreateMealHandler(logger *slog.Logger, templateFiles fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(
			templateFiles,
			"templates/layout.gohtml",
			"templates/navigation.gohtml",
			"templates/pages/meals/create_update.gohtml",
		)

		if err != nil {
			w.Write([]byte("Template error: " + err.Error()))

			return
		}

		formJson, err := getMessage(w, r, "formData")

		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"unable to fetch form data from cookie",
				slog.Any("err", err),
			)

			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		formData := meals.Meal{}

		if formJson != "" {
			json.Unmarshal([]byte(formJson), &formData)
		}

		type templateData struct {
			Title  string
			Form   meals.Meal
			Action string
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title:  "Create a Meal",
			Form:   formData,
			Action: "/meals/create",
		})
	})
}

// Helper to convert a form request into a Meal struct
func mealFromRequest(r http.Request) meals.Meal {
	r.ParseForm()

	var (
		id          int
		ingredients []meals.Ingredient
		tags        []meals.Tag
	)

	if r.Form.Has("id") {
		id, _ = strconv.Atoi(r.FormValue("id"))
	}

	for i := range len(r.Form["ingredientName"]) {
		var ingredient meals.Ingredient

		quantity, _ := strconv.Atoi(r.Form["ingredientQuantity"][i])

		id, _ := strconv.Atoi(r.Form["ingredientId"][i])

		ingredient.Id = id
		ingredient.Name = r.Form["ingredientName"][i]
		ingredient.Quantity = quantity
		ingredient.Unit = r.Form["ingredientUnit"][i]
		ingredients = append(ingredients, ingredient)
	}

	for i := range len(r.Form["tagName"]) {
		var tag meals.Tag

		id, _ := strconv.Atoi(r.Form["tagId"][i])

		tag.Id = id
		tag.Name = r.Form["tagName"][i]

		tags = append(tags, tag)
	}

	return meals.Meal{
		Id:          id,
		Name:        r.FormValue("name"),
		Notes:       r.FormValue("notes"),
		Ingredients: ingredients,
		Tags:        tags,
	}
}

// Handler to create a meal
//
// If the form is invalid, redirects back to the create a meal page.
// Redirects to the meal list page on success.
func PostMealHandler(logger *slog.Logger, service meals.MealService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := UserFromContext(r.Context())

		if user == nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"user missing from context",
			)

			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		meal := mealFromRequest(*r)

		meal.UserId = user.Id

		_, err := service.CreateMeal(r.Context(), &meal)

		if err != nil && errors.Is(err, meals.ErrMealFormInvalid) {
			formJson, _ := json.Marshal(meal)
			setMessage(w, "formData", string(formJson))

			http.Redirect(w, r, "/meals/create", http.StatusFound)

			return
		}

		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to create meal",
				slog.Any("err", err),
				slog.Int("userId", user.Id),
				slog.Any("meal", meal),
			)

			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		setMessage(w, "success", "Meal "+meal.Name+" has been created")

		http.Redirect(w, r, "/meals", http.StatusFound)
	})
}

// Handler to update a meal
//
// Only the owner of a meal can access this page.
// If a meal does not exist a 404 is returned.
// Expects the meal Id as a url path value with the name 'id'.
// If the form is invalid, redirects back to the edit a meal page.
// Redirects to the meal list page on success.
func PutMealHandler(logger *slog.Logger, service meals.MealService, repo meals.MealRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := UserFromContext(r.Context())

		if user == nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"user missing from context",
			)

			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		meal := mealFromRequest(*r)

		meal.Id, _ = strconv.Atoi(r.PathValue("id"))

		existingMeal, err := repo.Get(r.Context(), meal.Id)

		if err != nil {
			if errors.Is(err, meals.ErrMealNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)

				return
			}
		}

		if existingMeal.UserId != user.Id {
			http.Error(w, "You do not have permission to access this page", http.StatusForbidden)

			return
		}

		meal.UserId = user.Id

		err = service.UpdateMeal(r.Context(), &meal)

		if err != nil && errors.Is(err, meals.ErrMealFormInvalid) {
			formJson, _ := json.Marshal(meal)
			setMessage(w, "formData", string(formJson))

			http.Redirect(w, r, "/meals/"+r.PathValue("id"), http.StatusFound)

			return
		}

		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to create meal",
				slog.Any("err", err),
				slog.Int("userId", user.Id),
				slog.Any("meal", meal),
			)

			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		setMessage(w, "success", "Meal "+meal.Name+" has been updated")

		http.Redirect(w, r, "/meals", http.StatusFound)

	})
}

func PostDeleteMealHandler(logger *slog.Logger, repo meals.MealRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := UserFromContext(r.Context())

		if user == nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"user missing from context",
			)

			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		mealId, _ := strconv.Atoi(r.PathValue("id"))

		mealToDelete, err := repo.Get(r.Context(), mealId)

		if err != nil {
			if errors.Is(err, meals.ErrMealNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)

				return
			}
		}

		if mealToDelete.UserId != user.Id {
			http.Error(w, "You do not have permission to access this page", http.StatusForbidden)

			return
		}

		err = repo.Destroy(r.Context(), mealId)

		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to delete meal",
				slog.Any("err", err),
				slog.Int("userId", user.Id),
				slog.Int("mealId", mealId),
			)

			http.Error(w, "Server error", http.StatusInternalServerError)

			return
		}

		http.Redirect(w, r, "/meals", http.StatusFound)
	})
}

// API endpoint to search for existing ingredient names
//
// Returns JSON.
func GetSearchIngredientsHandler(logger *slog.Logger, metaDataRepo meals.MealMetaDataRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queryString := r.URL.Query().Get("query")

		results, err := metaDataRepo.FindIngredientNames(r.Context(), queryString)

		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to fetch ingredient names",
				slog.Any("err", err),
				slog.String("queryString", queryString),
			)

			http.Error(w, "server error", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(results)
	})
}

// API handler to search for tags by name
//
// Returns JSON.
func GetSearchTagHandler(logger *slog.Logger, metaDataRepo meals.MealMetaDataRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queryString := r.URL.Query().Get("query")

		results, err := metaDataRepo.FindTagNames(r.Context(), queryString)

		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to fetch tag names",
				slog.Any("err", err),
				slog.String("queryString", queryString),
			)

			http.Error(w, "server error", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(results)
	})
}

// API handler to search for units by name
//
// Returns JSON.
func GetSearchUnitHandler(logger *slog.Logger, metaDataRepo meals.MealMetaDataRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queryString := r.URL.Query().Get("query")

		results, err := metaDataRepo.FindUnitNames(r.Context(), queryString)

		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to fetch units",
				slog.Any("err", err),
				slog.String("queryString", queryString),
			)

			http.Error(w, "server error", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(results)
	})
}
