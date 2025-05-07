package web

import (
	"html/template"
	"io/fs"
	"net/http"
	"time"

	"github.com/jameswhoughton/meals/internal/account"
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

func GetPlannerHandler(templateFiles fs.FS, plannerRepo planner.Repository, accountRepo account.Repository) http.Handler {
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

		type day struct {
			Date  string
			Label string
			Meal  *planner.Meal
		}

		var startDate time.Time
		days := make([]day, 7)

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

		for i := range days {
			meal, _ := plannerRepo.Get(r.Context(), startDate, userId)

			day := day{
				Date:  startDate.Format("2006-01-02"),
				Label: startDate.Weekday().String(),
				Meal:  meal,
			}

			days[i] = day

			startDate = startDate.Add(24 * time.Hour)
		}

		type templateData struct {
			Title string
			Days  []day
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title: "Week Planner",
			Days:  days,
		})
	})
}
