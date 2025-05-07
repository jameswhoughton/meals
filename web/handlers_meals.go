package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/jameswhoughton/meals/internal/account"
	"github.com/jameswhoughton/meals/internal/meals"
)

// Render the meals list page.
//
// Only the authenticated user's meals are visible
// Results can be filtered
func GetMealsHandler(templateFiles fs.FS, repo meals.MealRepository) http.Handler {
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

		userId := r.Context().Value("userId").(int)
		queryString := r.URL.Query().Get("query")

		filter := meals.MealFilter{
			UserId: userId,
			Name:   &queryString,
		}

		results, err := repo.Find(r.Context(), filter)

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		type templateData struct {
			Title       string
			SearchQuery string
			Meals       []meals.Meal
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title:       "Meals",
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
func GetMealHandler(templateFiles fs.FS, repo meals.MealRepository) http.Handler {
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

		userId := r.Context().Value("userId").(int)
		mealId, _ := strconv.Atoi(r.PathValue("id"))

		meal, err := repo.Get(r.Context(), mealId)

		if err != nil {
			if errors.As(err, &meals.ErrorMealNotFound{Id: mealId}) {
				http.Error(w, err.Error(), http.StatusNotFound)

				return
			}
		}

		if meal.UserId != userId {
			http.Error(w, "You do not have permission to access this page", http.StatusForbidden)

			return
		}

		formJson, err := getMessage(w, r, "formData")

		if err != nil {
			log.Println(err)
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

// Render the create a meal page
func GetCreateMealHandler(templateFiles fs.FS) http.Handler {
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
			log.Println(err)
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
		ingredients []meals.MealIngredient
		tags        []meals.Tag
	)

	if r.Form.Has("id") {
		id, _ = strconv.Atoi(r.FormValue("id"))
	}

	mainIngredientIndex, err := strconv.Atoi(r.FormValue("isMain"))

	if err != nil {
		mainIngredientIndex = -1
	}

	for i := range len(r.Form["ingredientName"]) {
		var ingredient meals.MealIngredient

		quantity, _ := strconv.Atoi(r.Form["ingredientQuantity"][i])
		var isMain bool

		isMain = mainIngredientIndex == i

		id, _ := strconv.Atoi(r.Form["ingredientId"][i])

		ingredient.Id = id
		ingredient.Name = r.Form["ingredientName"][i]
		ingredient.Quantity = quantity
		ingredient.Unit = r.Form["ingredientUnit"][i]
		ingredient.IsMain = isMain
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
func PostMealHandler(service meals.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("userId").(int)

		meal := mealFromRequest(*r)

		meal.UserId = userId

		_, err := service.CreateMeal(r.Context(), &meal)

		if err != nil && errors.Is(err, meals.ErrorFormInvalid{}) {
			formJson, _ := json.Marshal(meal)
			setMessage(w, "formData", string(formJson))

			http.Redirect(w, r, "/meals/create", http.StatusFound)

			return
		}

		if err != nil {
			log.Println(err)
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
func PutMealHandler(service meals.Service, repo meals.MealRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("userId").(int)

		meal := mealFromRequest(*r)

		meal.Id, _ = strconv.Atoi(r.PathValue("id"))

		existingMeal, err := repo.Get(r.Context(), meal.Id)

		if err != nil {
			if errors.As(err, &meals.ErrorMealNotFound{Id: meal.Id}) {
				http.Error(w, err.Error(), http.StatusNotFound)

				return
			}
		}

		if existingMeal.UserId != userId {
			http.Error(w, "You do not have permission to access this page", http.StatusForbidden)

			return
		}

		meal.UserId = userId

		err = service.UpdateMeal(r.Context(), &meal)

		if err != nil && errors.Is(err, meals.ErrorFormInvalid{}) {
			formJson, _ := json.Marshal(meal)
			setMessage(w, "formData", string(formJson))

			http.Redirect(w, r, "/meals/"+r.PathValue("id"), http.StatusFound)

			return
		}

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		setMessage(w, "success", "Meal "+meal.Name+" has been updated")

		http.Redirect(w, r, "/meals", http.StatusFound)

	})
}

func PostDeleteMealHandler(repo meals.MealRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("userId").(int)

		mealId, _ := strconv.Atoi(r.PathValue("id"))

		mealToDelete, err := repo.Get(r.Context(), mealId)

		if err != nil {
			if errors.As(err, &meals.ErrorMealNotFound{Id: mealId}) {
				http.Error(w, err.Error(), http.StatusNotFound)

				return
			}
		}

		if mealToDelete.UserId != userId {
			http.Error(w, "You do not have permission to access this page", http.StatusForbidden)

			return
		}

		err = repo.Destroy(r.Context(), mealId)

		if err != nil {
			http.Error(w, "Server error: Unable to delete meal", http.StatusInternalServerError)

			return
		}

		http.Redirect(w, r, "/meals", http.StatusFound)
	})
}

// Render the Planner page

func calculateStartDate(date time.Time, startDay int) time.Time {
	if date.Weekday() == time.Weekday(startDay) {
		return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	}

	diff := (date.Weekday() - time.Weekday(startDay))

	if diff < 0 {
		diff += 7
	}

	date = date.Add(-time.Duration(diff) * 24 * time.Hour)

	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
}

func GetPlannerHandler(templateFiles fs.FS, mealService meals.Service, accountRepo account.Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(
			templateFiles,
			"templates/layout.gohtml",
			"templates/navigation.gohtml",
			"templates/pages/planner.gohtml",
		)

		if err != nil {
			w.Write([]byte("Template error: " + err.Error()))

			return
		}
		userId := r.Context().Value("userId").(int)

		user, err := accountRepo.Get(r.Context(), account.GetForm{Id: &userId})

		if err != nil {

		}

		type Day struct {
			Date  string
			Label string
			Meal  *meals.Meal
		}

		var startDate time.Time
		var days []Day

		if r.PathValue("date") == "" {
			startDate = calculateStartDate(time.Now(), user.MealStartDay)
		} else {
			parsedDate, err := time.Parse("02-01-2006", r.PathValue("date"))

			if err != nil {
				http.Error(w, "Cannot parse date: "+r.PathValue("date"), http.StatusBadRequest)

				return
			}

			startDate = calculateStartDate(parsedDate, user.MealStartDay)
		}

		fmt.Println(startDate)

		type templateData struct {
			Title string
			Days  []Day
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title: "Week Planner",
			Days:  days,
		})
	})
}

// Render the ingredients list page
//
// Ingredients can be filtered by name.
// Only ingredients owned by the authed user are visible.
func GetIngredientsHandler(templateFiles fs.FS, ingredients meals.IngredientRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(
			templateFiles,
			"templates/layout.gohtml",
			"templates/navigation.gohtml",
			"templates/pages/ingredients/list.gohtml",
		)

		if err != nil {
			w.Write([]byte("Template error: " + err.Error()))

			return
		}

		userId := r.Context().Value("userId").(int)
		queryString := r.URL.Query().Get("query")

		ingredients, err := ingredients.Find(r.Context(), queryString, userId)

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		type templateData struct {
			Title       string
			SearchQuery string
			Ingredients []meals.Ingredient
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title:       "Ingredients",
			SearchQuery: queryString,
			Ingredients: ingredients,
		})
	})
}

// API endpoint to search for ingredient by name

// Results are limited to the authenticated user.
// Returns JSON.
func GetSearchIngredientsHandler(ingredients meals.IngredientRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("userId").(int)

		queryString := r.URL.Query().Get("query")

		results, err := ingredients.Find(r.Context(), queryString, userId)

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(results)
	})
}

// Render the edit ingredient page
//
// Redirects to a 404 page if the ingredient does not exist.
// Only the owner of an ingredient can access the page.
// Expects the ingredient id to be provided as a path value with the name 'id'.
func GetIngredientHandler(templateFiles fs.FS, ingredients meals.IngredientRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(
			templateFiles,
			"templates/layout.gohtml",
			"templates/navigation.gohtml",
			"templates/pages/ingredients/edit.gohtml",
		)

		if err != nil {
			w.Write([]byte("Template error: " + err.Error()))

			return
		}

		userId := r.Context().Value("userId").(int)
		ingredientId, _ := strconv.Atoi(r.PathValue("id"))

		ingredient, err := ingredients.GetById(r.Context(), ingredientId)

		if err != nil {
			if errors.As(err, &meals.ErrorIngredientNotFound{Id: ingredientId}) {
				http.Error(w, err.Error(), http.StatusNotFound)

				return
			}
		}

		if ingredient.UserId != userId {
			http.Error(w, "You do not have permission to access this page", http.StatusForbidden)

			return
		}

		formJson, err := getMessage(w, r, "formData")

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		if formJson != "" {
			json.Unmarshal([]byte(formJson), &ingredient)
		}

		type templateData struct {
			Title string
			Form  meals.Ingredient
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title: "Edit Ingredients",
			Form:  ingredient,
		})
	})
}

// Handler to update an ingredient
//
// An ingredient can only be updated by it's owner.
// Returns a 404 if the ingredient does not exist.
// Redirects back to the ingredients list page on success.
// If the form is invalid, redirect back to the edit ingredient page.
func PutIngredientHandler(service meals.Service, repo meals.IngredientRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		userId := r.Context().Value("userId").(int)

		ingredientId, _ := strconv.Atoi(r.PathValue("id"))

		existingIngredient, err := repo.GetById(r.Context(), ingredientId)

		if err != nil {
			if errors.As(err, &meals.ErrorIngredientNotFound{Id: ingredientId}) {
				http.Error(w, err.Error(), http.StatusNotFound)

				return
			}
		}

		if existingIngredient.UserId != userId {
			http.Error(w, "You do not have permission to access this page", http.StatusForbidden)

			return
		}

		ingredient := meals.Ingredient{
			Id:     ingredientId,
			UserId: userId,
			Name:   r.FormValue("name"),
		}

		err = service.UpdateIngredient(r.Context(), &ingredient)

		if err != nil && errors.Is(err, meals.ErrorFormInvalid{}) {
			formJson, _ := json.Marshal(ingredient)
			setMessage(w, "formData", string(formJson))

			http.Redirect(w, r, "/ingredients/"+r.PathValue("id"), http.StatusFound)

			return
		}

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		http.Redirect(w, r, "/ingredients", http.StatusFound)
	})
}

// API handler to search for tags by name
//
// Limited to tags belonging to the authed user.
// Returns JSON.
func GetSearchTagHandler(tags meals.TagRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("userId").(int)

		queryString := r.URL.Query().Get("query")

		results, err := tags.Find(r.Context(), queryString, userId)

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(results)
	})
}

// Render the tags list page
//
// Only lists tags owned by the authed user.
func GetTagsHandler(templateFiles fs.FS, tags meals.TagRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(
			templateFiles,
			"templates/layout.gohtml",
			"templates/navigation.gohtml",
			"templates/pages/tags/list.gohtml",
		)

		if err != nil {
			w.Write([]byte("Template error: " + err.Error()))

			return
		}

		userId := r.Context().Value("userId").(int)
		queryString := r.URL.Query().Get("query")

		tags, err := tags.Find(r.Context(), queryString, userId)

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		type templateData struct {
			Title       string
			SearchQuery string
			Tags        []meals.Tag
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title:       "Tags",
			SearchQuery: queryString,
			Tags:        tags,
		})
	})
}

// Render the edit tag page
//
// Returns 404 if the tag does not exist
// Only the owner of the tag can access the page
// Expects the tag id to be set as a path value with the name 'id'
func GetTagHandler(templateFiles fs.FS, tags meals.TagRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(
			templateFiles,
			"templates/layout.gohtml",
			"templates/navigation.gohtml",
			"templates/pages/tags/edit.gohtml",
		)

		if err != nil {
			w.Write([]byte("Template error: " + err.Error()))

			return
		}

		userId := r.Context().Value("userId").(int)
		tagId, _ := strconv.Atoi(r.PathValue("id"))

		tag, err := tags.GetById(r.Context(), tagId)

		if err != nil {
			if errors.As(err, &meals.ErrorTagNotFound{Id: tagId}) {
				http.Error(w, err.Error(), http.StatusNotFound)

				return
			}
		}

		if tag.UserId != userId {
			http.Error(w, "You do not have permission to access this page", http.StatusForbidden)

			return
		}

		formJson, err := getMessage(w, r, "formData")

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		if formJson != "" {
			json.Unmarshal([]byte(formJson), &tag)
		}

		type templateData struct {
			Title string
			Form  meals.Tag
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title: "Edit Tag",
			Form:  tag,
		})
	})
}

// Handler to update tag
//
// Only the owner of a tag can edit it
// On success redirects to the tag list page
// If there are validation errors, redirects to the edit tag page
func PutTagHandler(service meals.Service, repo meals.TagRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		userId := r.Context().Value("userId").(int)

		tagId, _ := strconv.Atoi(r.PathValue("id"))

		existingTag, err := repo.GetById(r.Context(), tagId)

		if err != nil {
			if errors.As(err, &meals.ErrorTagNotFound{Id: tagId}) {
				http.Error(w, err.Error(), http.StatusNotFound)

				return
			}
		}

		if existingTag.UserId != userId {
			http.Error(w, "You do not have permission to access this page", http.StatusForbidden)

			return
		}

		tag := meals.Tag{
			Id:     tagId,
			UserId: userId,
			Name:   r.FormValue("name"),
		}

		err = service.UpdateTag(r.Context(), &tag)

		if err != nil && errors.Is(err, meals.ErrorFormInvalid{}) {
			formJson, _ := json.Marshal(tag)
			setMessage(w, "formData", string(formJson))

			http.Redirect(w, r, "/tags/"+r.PathValue("id"), http.StatusFound)

			return
		}

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		http.Redirect(w, r, "/tags", http.StatusFound)
	})
}
