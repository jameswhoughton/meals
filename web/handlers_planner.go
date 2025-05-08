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
			Date  string
			Label string
			Meal  planner.Meal
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
				Date:  date.Format("2006-01-02"),
				Label: dateLabel(date),
				Meal:  meal,
			}

			days[i] = day

			date = date.Add(24 * time.Hour)
		}

		type templateData struct {
			Title    string
			Days     []day
			Previous string
			Next     string
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title:    "Week Planner",
			Days:     days,
			Previous: startDate.Add(-time.Duration(7*24) * time.Hour).Format("2006-01-02"),
			Next:     startDate.Add(time.Duration(7*24) * time.Hour).Format("2006-01-02"),
		})
	})
}

func GetEditDayHandler(templateFiles fs.FS, plannerRepo planner.Repository, mealRepo meals.MealRepository, tagRepo meals.TagRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contains := func(tags []int, id int) bool {
			return slices.Contains(tags, id)
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
		filterTags := make([]int, len(r.Form["tags"]))
		for i := range len(r.Form["tags"]) {
			filterTags[i], _ = strconv.Atoi(r.Form["tags"][i])
		}
		filter := meals.MealFilter{
			UserId:  userId,
			Name:    &filterSearch,
			HasTags: filterTags,
		}

		filteredMeals, err := mealRepo.Find(r.Context(), filter)

		tags, err := tagRepo.Find(r.Context(), "", userId)

		type templateData struct {
			Title        string
			Date         string
			Meal         planner.Meal
			Meals        []meals.Meal
			Tags         []meals.Tag
			FilterSearch string
			FilterTags   []int
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title:        "Editing " + dateLabel(parsedDate),
			Date:         parsedDate.Format("2006-01-02"),
			Meal:         meal,
			Meals:        filteredMeals,
			Tags:         tags,
			FilterSearch: filterSearch,
			FilterTags:   filterTags,
		})
	})
}
