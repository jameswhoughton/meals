package meals

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/jameswhoughton/meals/web/helpers"
)

func GetMealsHandler(templateFiles fs.FS, meals MealRepository) http.Handler {
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

		filter := MealFilter{
			UserId: userId,
			Name:   &queryString,
		}

		meals, err := meals.Find(filter)

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		type templateData struct {
			Title       string
			SearchQuery string
			Meals       []Meal
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title:       "Meals",
			SearchQuery: queryString,
			Meals:       meals,
		})
	})
}

func GetMealHandler(templateFiles fs.FS, meals MealRepository) http.Handler {
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

		meal, err := meals.Get(mealId)

		if err != nil {
			if errors.As(err, &ErrorMealNotFound{Id: mealId}) {
				http.Error(w, err.Error(), http.StatusNotFound)

				return
			}
		}

		if meal.UserId != userId {
			http.Error(w, "You do not have permission to access this page", http.StatusForbidden)

			return
		}

		formJson, err := helpers.GetMessage(w, r, "formData")

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
			Form   Meal
			Action string
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title:  meal.Name,
			Form:   formData,
			Action: "/meals/" + r.PathValue("id"),
		})
	})
}

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

		formJson, err := helpers.GetMessage(w, r, "formData")

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		formData := Meal{}

		if formJson != "" {
			json.Unmarshal([]byte(formJson), &formData)
		}

		type templateData struct {
			Title  string
			Form   Meal
			Action string
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title:  "Create a Meal",
			Form:   formData,
			Action: "/meals/create",
		})
	})
}

func mealFromRequest(r http.Request) Meal {
	r.ParseForm()

	var (
		id          int
		ingredients []MealIngredient
	)

	if r.Form.Has("id") {
		id, _ = strconv.Atoi(r.FormValue("id"))
	}

	mainIngredientIndex, err := strconv.Atoi(r.FormValue("isMain"))

	if err != nil {
		mainIngredientIndex = -1
	}

	for i := range len(r.Form["ingredientName"]) {
		var ingredient MealIngredient

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

	return Meal{
		Id:          id,
		Name:        r.FormValue("name"),
		Notes:       r.FormValue("notes"),
		Ingredients: ingredients,
	}
}

func PostMealHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("userId").(int)

		meal := mealFromRequest(*r)

		meal.UserId = userId

		_, err := service.CreateMeal(&meal)

		if err != nil && errors.Is(err, ErrorFormInvalid{}) {
			formJson, _ := json.Marshal(meal)
			helpers.SetMessage(w, "formData", string(formJson))

			http.Redirect(w, r, "/meals/create", http.StatusFound)

			return
		}

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		helpers.SetMessage(w, "success", "Meal "+meal.Name+" has been created")

		http.Redirect(w, r, "/meals", http.StatusFound)
	})
}

func PutMealHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("userId").(int)

		meal := mealFromRequest(*r)

		meal.Id, _ = strconv.Atoi(r.PathValue("id"))

		existingMeal, err := service.meals.Get(meal.Id)

		if err != nil {
			if errors.As(err, &ErrorMealNotFound{Id: meal.Id}) {
				http.Error(w, err.Error(), http.StatusNotFound)

				return
			}
		}

		if existingMeal.UserId != userId {
			http.Error(w, "You do not have permission to access this page", http.StatusForbidden)

			return
		}

		meal.UserId = userId

		err = service.UpdateMeal(&meal)

		if err != nil && errors.Is(err, ErrorFormInvalid{}) {
			formJson, _ := json.Marshal(meal)
			helpers.SetMessage(w, "formData", string(formJson))

			http.Redirect(w, r, "/meals/"+r.PathValue("id"), http.StatusFound)

			return
		}

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		helpers.SetMessage(w, "success", "Meal "+meal.Name+" has been updated")

		http.Redirect(w, r, "/meals", http.StatusFound)

	})
}

// DeleteMealHandler

// GetPlannerHandler

func GetIngredientsHandler(templateFiles fs.FS, ingredients IngredientRepository) http.Handler {
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

		ingredients, err := ingredients.Find(queryString, userId)

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		type templateData struct {
			Title       string
			SearchQuery string
			Ingredients []Ingredient
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title:       "Ingredients",
			SearchQuery: queryString,
			Ingredients: ingredients,
		})
	})
}

func GetSearchIngredientsHandler(ingredients IngredientRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("userId").(int)

		queryString := r.URL.Query().Get("query")

		ingredients, err := ingredients.Find(queryString, userId)

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(ingredients)
	})
}

func GetIngredientHandler(templateFiles fs.FS, ingredients IngredientRepository) http.Handler {
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

		ingredient, err := ingredients.GetById(ingredientId)

		if err != nil {
			if errors.As(err, &ErrorIngredientNotFound{Id: ingredientId}) {
				http.Error(w, err.Error(), http.StatusNotFound)

				return
			}
		}

		if ingredient.UserId != userId {
			http.Error(w, "You do not have permission to access this page", http.StatusForbidden)

			return
		}

		formJson, err := helpers.GetMessage(w, r, "formData")

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
			Form  Ingredient
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title: "Edit Ingredients",
			Form:  ingredient,
		})
	})
}

func PutIngredientHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		userId := r.Context().Value("userId").(int)

		ingredientId, _ := strconv.Atoi(r.PathValue("id"))

		ingredient := Ingredient{
			Id:     ingredientId,
			UserId: userId,
			Name:   r.FormValue("name"),
		}

		err := service.UpdateIngredient(&ingredient)

		if err != nil && errors.Is(err, ErrorFormInvalid{}) {
			formJson, _ := json.Marshal(ingredient)
			helpers.SetMessage(w, "formData", string(formJson))

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
