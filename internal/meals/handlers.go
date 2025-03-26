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
		quick       bool
		family      bool
		easy        bool
		ingredients []MealIngredient
	)

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

		ingredient.Name = r.Form["ingredientName"][i]
		ingredient.Quantity = quantity
		ingredient.Unit = r.Form["ingredientUnit"][i]
		ingredient.IsMain = isMain
		ingredients = append(ingredients, ingredient)
	}

	return Meal{
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
