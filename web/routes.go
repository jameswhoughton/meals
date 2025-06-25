package web

import (
	"embed"
	"errors"
	"log"
	"log/slog"
	"net/http"

	"github.com/jameswhoughton/meals/internal/account"
	"github.com/jameswhoughton/meals/internal/meals"
	"github.com/jameswhoughton/meals/internal/planner"
)

//go:embed templates/*.gohtml
//go:embed templates/**/**/*.gohtml
//go:embed templates/**/**/**/*.gohtml
var templateFiles embed.FS

//go:embed static/*
var publicFiles embed.FS

func AddRoutes(
	mux *http.ServeMux,
	logger *slog.Logger,
	accountService account.Service,
	mealService meals.Service,
	sessionService SessionService,
	plannerService planner.Service,
	accountRepository account.Repository,
	mealRepository meals.Repository,
	sessionRepository SessionRepository,
	plannerRepository planner.Repository,
) {
	// Middleware
	isAuthed := NewIsAuthenticatedMiddleware(accountService, sessionService)

	// Root redirect
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("session")

		if err != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				http.Redirect(w, r, "/login", http.StatusFound)
			default:
				log.Println(err)
				http.Error(w, "server error", http.StatusInternalServerError)
			}
			return
		}

		http.Redirect(w, r, "/account", http.StatusFound)

	})

	// Static files
	mux.Handle("GET /static/", getStaticFilesHandler(publicFiles))

	// Authentication
	mux.Handle("GET /login", GetLoginHandler(logger, templateFiles))
	mux.Handle("POST /login", PostLoginHandler(logger, accountService, sessionService))
	mux.Handle("GET /register", GetRegistrationHandler(logger, templateFiles))
	mux.Handle("POST /register", PostRegistrationHandler(logger, accountService))
	mux.Handle("GET /logout", GetLogoutHandler(logger, accountService, sessionRepository))

	// Account
	mux.Handle("GET /account", isAuthed(GetAccountHandler(logger, templateFiles, sessionService)))
	mux.Handle("POST /account", isAuthed(PutAccountHandler(logger, accountService, sessionService)))

	// Meals
	mux.Handle("GET /meals/create", isAuthed(GetCreateMealHandler(logger, templateFiles)))
	mux.Handle("POST /meals/create", isAuthed(PostMealHandler(logger, mealService)))
	mux.Handle("GET /meals", isAuthed(GetMealsHandler(logger, templateFiles, mealRepository)))
	mux.Handle("GET /meals/{id}", isAuthed(GetMealHandler(logger, templateFiles, mealRepository)))
	mux.Handle("GET /meals/{id}/edit", isAuthed(GetMealEditHandler(logger, templateFiles, mealRepository)))
	mux.Handle("POST /meals/{id}", isAuthed(PutMealHandler(logger, mealService, mealRepository)))
	mux.Handle("POST /meals/{id}/delete", isAuthed(PostDeleteMealHandler(logger, mealRepository)))

	// Planner
	mux.Handle("GET /planner", isAuthed(GetPlannerHandler(logger, templateFiles, plannerService, accountRepository)))
	mux.Handle("GET /planner/{date}", isAuthed(GetEditDayHandler(logger, templateFiles, plannerRepository, mealRepository)))
	mux.Handle("POST /planner/{date}", isAuthed(PostEditDayHandler(logger, plannerRepository, mealRepository)))
	mux.Handle("GET /planner/{date}/ingredients", isAuthed(GetPlannedIngredientsHandler(logger, plannerService, accountRepository)))

	// API
	mux.Handle("GET /api/ingredients", isAuthed(GetSearchIngredientsHandler(logger, mealRepository)))
	mux.Handle("GET /api/tags", isAuthed(GetSearchTagHandler(logger, mealRepository)))
	mux.Handle("GET /api/units", isAuthed(GetSearchUnitHandler(logger, mealRepository)))
}
