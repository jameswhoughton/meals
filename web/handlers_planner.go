package web

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/jameswhoughton/meals/internal/account"
	"github.com/jameswhoughton/meals/internal/meals"
	"github.com/jameswhoughton/meals/internal/planner"
)

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

func dateLabel(date time.Time) string {
	return fmt.Sprintf("%s (%s)", date.Weekday().String(), date.Format("02/01"))
}

func GetPlannerHandler(templateFiles fs.FS, plannerRepo planner.Repository, accountRepo account.Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(
			templateFiles,
			"templates/layout.gohtml",
			"templates/navigation.gohtml",
			"templates/pages/planner/index.gohtml",
		)
		if err != nil {
			w.Write([]byte("Template error: " + err.Error()))

			return
		}
		userId := r.Context().Value("userId").(int)

		user, err := accountRepo.Get(r.Context(), account.GetForm{Id: &userId})

		if err != nil {

		}

		type day struct {
			Date      string
			Label     string
			Meal      planner.Meal
			IsWeekend bool
			IsToday   bool
		}

		var startDate time.Time
		days := make([]day, 7)

		dateFromUrl := r.FormValue("date")

		if dateFromUrl == "" {
			startDate = calculateStartDate(time.Now(), user.MealStartDay)
		} else {
			parsedDate, err := time.Parse("2006-01-02", dateFromUrl)

			if err != nil {
				http.Error(w, "Cannot parse date: "+dateFromUrl, http.StatusBadRequest)

				return
			}

			startDate = calculateStartDate(parsedDate, user.MealStartDay)
		}

		date := startDate

		for i := range days {
			meal, _ := plannerRepo.Get(r.Context(), date, userId)

			day := day{
				Date:      date.Format("2006-01-02"),
				Label:     dateLabel(date),
				Meal:      meal,
				IsWeekend: slices.Contains([]string{time.Saturday.String(), time.Sunday.String()}, date.Weekday().String()),
				IsToday:   date.Format("2006-01-02") == time.Now().Format("2006-01-02"),
			}

			days[i] = day

			date = date.Add(24 * time.Hour)
		}

		type templateData struct {
			Title      string
			Days       []day
			Previous   string
			Next       string
			ChosenDate string
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title:      "Week Planner",
			Days:       days,
			Previous:   startDate.Add(-time.Duration(7*24) * time.Hour).Format("2006-01-02"),
			Next:       startDate.Add(time.Duration(7*24) * time.Hour).Format("2006-01-02"),
			ChosenDate: startDate.Format("2006-01-02"),
		})
	})
}

func GetEditDayHandler(templateFiles fs.FS, plannerRepo planner.Repository, mealRepo meals.Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contains := func(tags []string, name string) bool {
			return slices.Contains(tags, name)
		}

		funcMap := template.FuncMap{
			"contains": contains,
		}

		tmpl := template.New("edit-day").Funcs(funcMap)

		tmpl, err := tmpl.ParseFS(
			templateFiles,
			"templates/layout.gohtml",
			"templates/navigation.gohtml",
			"templates/pages/planner/edit.gohtml",
		)

		if err != nil {
			w.Write([]byte("Template error: " + err.Error()))

			return
		}

		userId := r.Context().Value("userId").(int)

		parsedDate, err := time.Parse("2006-01-02", r.PathValue("date"))

		meal, _ := plannerRepo.Get(r.Context(), parsedDate, userId)

		// Parse any filter params
		r.ParseForm()

		filterSearch := r.Form.Get("query")
		filterTags := make([]string, len(r.Form["tags"]))
		for i := range len(r.Form["tags"]) {
			filterTags[i] = r.Form["tags"][i]
		}
		filter := meals.MealFilter{
			UserId: userId,
			Name:   &filterSearch,
			Tags:   filterTags,
		}

		filteredMeals, err := mealRepo.Find(r.Context(), filter)

		tags, err := mealRepo.TagNamesForUser(r.Context(), userId)

		type templateData struct {
			Title        string
			Date         string
			Meal         planner.Meal
			Meals        []meals.Meal
			Tags         []string
			FilterSearch string
			FilterTags   []string
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title:        dateLabel(parsedDate),
			Date:         parsedDate.Format("2006-01-02"),
			Meal:         meal,
			Meals:        filteredMeals,
			Tags:         tags,
			FilterSearch: filterSearch,
			FilterTags:   filterTags,
		})
	})
}

func PostEditDayHandler(plannerRepo planner.Repository, mealRepo meals.Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("userId").(int)

		parsedDate, err := time.Parse("2006-01-02", r.PathValue("date"))

		if err != nil {
			http.Error(w, "Invalid date", http.StatusBadRequest)

			return
		}

		mealId := r.FormValue("meal_id")
		clearMeal := r.FormValue("action") == "clear"

		plannerRepo.Clear(r.Context(), parsedDate, userId)

		if !clearMeal && mealId != "" {
			mealId, err := strconv.Atoi(mealId)

			if err != nil {
				http.Error(w, "Meal ID must be an integer", http.StatusBadRequest)

				return
			}

			meal, err := mealRepo.Get(r.Context(), mealId)

			if err != nil {
				http.Error(w, "Meal not found", http.StatusNotFound)

				return
			}

			if meal.UserId != userId {
				http.Error(w, "You do not have permission to assign this meal", http.StatusForbidden)
			}

			err = plannerRepo.Add(r.Context(), parsedDate, mealId)

			if err != nil {
				http.Error(w, "Server error: unable to save meal to date", http.StatusInternalServerError)

				return
			}
		}

		http.Redirect(w, r, "/planner?date="+r.PathValue("date"), http.StatusFound)
	})
}

func GetPlannedIngredientsHandler(plannerSerivce planner.Service, accountRepo account.Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(
			templateFiles,
			"templates/layout.gohtml",
			"templates/navigation.gohtml",
			"templates/pages/planner/ingredients.gohtml",
		)

		if err != nil {
			w.Write([]byte("Template error: " + err.Error()))

			return
		}

		userId := r.Context().Value("userId").(int)
		parsedDate, err := time.Parse("2006-01-02", r.PathValue("date"))

		if err != nil {
			http.Error(w, "Invalid date", http.StatusBadRequest)

			return
		}

		user, err := accountRepo.Get(r.Context(), account.GetForm{Id: &userId})

		startDate := calculateStartDate(parsedDate, user.MealStartDay)

		// At the momeent we limit to a 7 day window, this could in future be user configurable.
		endDate := startDate.AddDate(0, 0, 7)

		ingredients, err := plannerSerivce.GetIngredients(r.Context(), startDate, endDate, userId)

		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)

			return
		}

		type templateData struct {
			Title       string
			Ingredients []planner.Ingredient
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title:       fmt.Sprintf("Ingredients for week %s - %s", startDate.Format("02/01"), endDate.Format("02/01")),
			Ingredients: ingredients,
		})
	})
}
