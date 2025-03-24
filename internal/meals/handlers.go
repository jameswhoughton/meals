package meals

import (
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
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

// PostMealHandler

// PutMealHandler

// DeleteMealHandler

// GetPlannerHandler
