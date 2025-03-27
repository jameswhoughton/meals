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

// GetMealsHandler

// GetMealHandler

// GetCreateMealHandler
func GetCreateMealHandler(templateFiles fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(
			templateFiles,
			"templates/layout.gohtml",
			"templates/navigation.gohtml",
			"templates/pages/meals/create.gohtml",
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
			Title string
			Form  Meal
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title: "Create Meal",
			Form:  formData,
		})
	})
}

func mealFromRequest(r http.Request) Meal {
	r.ParseForm()

	var (
		id          int
		quick       bool
		family      bool
		easy        bool
		ingredients []MealIngredient
	)

	if r.Form.Has("id") {
		id, _ = strconv.Atoi(r.FormValue("id"))
	}

	if r.FormValue("quick") == "1" {
		quick = true
	}

	if r.FormValue("easy") == "1" {
		easy = true
	}

	if r.FormValue("family") == "1" {
		family = true
	}
	mainIngredientIndex, err := strconv.Atoi(r.FormValue("isMain"))

	if err != nil {
		mainIngredientIndex = -1
	}

	for i := range len(r.Form["ingredientName"]) {
		var ingredient MealIngredient

		quantity, _ := strconv.Atoi(r.Form["ingredientQuantity"][i])
		var isMain bool

		if mainIngredientIndex == i {
			isMain = true
		}

		id, _ := strconv.Atoi(r.Form["ingredientId"][i])

		ingredient.Id = id
		ingredient.Name = r.Form["ingredientName"][i]
		ingredient.Quantity = quantity
		ingredient.Unit = r.Form["ingredientUnit"][i]
		ingredient.IsMain = isMain
		ingredients = append(ingredients, ingredient)
	}

	return Meal{
		Id:    id,
		Name:  r.FormValue("name"),
		Notes: r.FormValue("notes"),
		Attributes: MealAttributes{
			Quick:  quick,
			Family: family,
			Easy:   easy,
		},
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

// PutMealHandler

// DeleteMealHandler

// GetPlannerHandler

func GetIngredientsHandler(templateFiles fs.FS, service Service) http.Handler {
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

		ingredients, err := service.repo.FindIngredients(queryString, userId)

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

func GetSearchIngredientsHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("userId").(int)

		queryString := r.URL.Query().Get("query")

		ingredients, err := service.ListIngredients(queryString, userId)

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(ingredients)
	})
}

func GetIngredientHandler(templateFiles fs.FS, serivce Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}
