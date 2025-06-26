package web

import (
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/jameswhoughton/meals"
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

func GetPlannerHandler(logger *slog.Logger, templateFiles fs.FS, plannerService planner.Service, accountRepo meals.UserRepository) http.Handler {
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

		var startDate time.Time

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

		days, err := plannerService.GetMeals(r.Context(), startDate, 7, user.Id)

		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"unable to fetch days",
				slog.Any("err", err),
				slog.Int("userId", user.Id),
				slog.String("date", startDate.String()),
			)

			http.Error(w, "Server error", http.StatusInternalServerError)

			return
		}

		type templateData struct {
			Title      string
			Days       []planner.Day
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

func GetEditDayHandler(logger *slog.Logger, templateFiles fs.FS, plannerRepo planner.Repository, mealRepo meals.MealRepository) http.Handler {
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

		parsedDate, err := time.Parse("2006-01-02", r.PathValue("date"))

		meal, _ := plannerRepo.Get(r.Context(), parsedDate, user.Id)

		// Parse any filter params
		r.ParseForm()

		filterSearch := r.Form.Get("query")
		filterTags := make([]string, len(r.Form["tags"]))
		for i := range len(r.Form["tags"]) {
			filterTags[i] = r.Form["tags"][i]
		}
		filter := meals.MealFilter{
			UserId: user.Id,
			Name:   &filterSearch,
			Tags:   filterTags,
		}

		filteredMeals, err := mealRepo.Find(r.Context(), filter)

		tags, err := mealRepo.TagNamesForUser(r.Context(), user.Id)

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
			Title:        fmt.Sprintf("%s (%s)", parsedDate.Weekday().String(), parsedDate.Format("02/01")),
			Date:         parsedDate.Format("2006-01-02"),
			Meal:         meal,
			Meals:        filteredMeals,
			Tags:         tags,
			FilterSearch: filterSearch,
			FilterTags:   filterTags,
		})
	})
}

func PostEditDayHandler(logger *slog.Logger, plannerRepo planner.Repository, mealRepo meals.MealRepository) http.Handler {
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

		parsedDate, err := time.Parse("2006-01-02", r.PathValue("date"))

		if err != nil {
			http.Error(w, "Invalid date", http.StatusBadRequest)

			return
		}

		mealId := r.FormValue("meal_id")
		clearMeal := r.FormValue("action") == "clear"

		plannerRepo.Clear(r.Context(), parsedDate, user.Id)

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

			if meal.UserId != user.Id {
				http.Error(w, "You do not have permission to assign this meal", http.StatusForbidden)
			}

			err = plannerRepo.Add(r.Context(), parsedDate, mealId)

			if err != nil {
				logger.LogAttrs(
					r.Context(),
					slog.LevelError,
					"unable to save meal to date",
					slog.Int("userId", user.Id),
					slog.Int("mealId", mealId),
					slog.String("date", parsedDate.String()),
				)

				http.Error(w, "Server error", http.StatusInternalServerError)

				return
			}
		}

		http.Redirect(w, r, "/planner?date="+r.PathValue("date"), http.StatusFound)
	})
}

func GetPlannedIngredientsHandler(logger *slog.Logger, plannerSerivce planner.Service) http.Handler {
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
		parsedDate, err := time.Parse("2006-01-02", r.PathValue("date"))

		if err != nil {
			http.Error(w, "Invalid date", http.StatusBadRequest)

			return
		}

		startDate := calculateStartDate(parsedDate, user.MealStartDay)

		// At the momeent we limit to a 7 day window, this could in future be user configurable.
		endDate := startDate.AddDate(0, 0, 7)

		ingredients, err := plannerSerivce.GetIngredients(r.Context(), startDate, endDate, user.Id)

		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"unable to get ingredients list",
				slog.Int("userId", user.Id),
				slog.String("startDate", startDate.String()),
				slog.String("endDate", endDate.String()),
			)

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
