package web

import (
	"embed"
	"errors"
	"log"
	"log/slog"
	"net/http"

	"github.com/jameswhoughton/meals"
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
	userService meals.UserService,
	mealService meals.MealService,
	sessionService SessionService,
	plannerService planner.Service,
	userRepository meals.UserRepository,
	mealRepository meals.MealRepository,
	sessionRepository SessionRepository,
	plannerRepository planner.Repository,
) {
	// Middleware
	isAuthed := NewIsAuthenticatedMiddleware(userService, sessionService)

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
	mux.Handle("POST /login", PostLoginHandler(logger, userService, sessionService))
	mux.Handle("GET /register", GetRegistrationHandler(logger, templateFiles))
	mux.Handle("POST /register", PostRegistrationHandler(logger, userService))
	mux.Handle("GET /logout", GetLogoutHandler(logger, userService, sessionRepository))

	// Account
	mux.Handle("GET /account", isAuthed(GetAccountHandler(logger, templateFiles, sessionService)))
	mux.Handle("POST /account", isAuthed(PutAccountHandler(logger, userService, sessionService)))

	// Meals
	mux.Handle("GET /meals/create", isAuthed(GetCreateMealHandler(logger, templateFiles)))
	mux.Handle("POST /meals/create", isAuthed(PostMealHandler(logger, mealService)))
	mux.Handle("GET /meals", isAuthed(GetMealsHandler(logger, templateFiles, mealRepository)))
	mux.Handle("GET /meals/{id}", isAuthed(GetMealHandler(logger, templateFiles, mealRepository)))
	mux.Handle("GET /meals/{id}/edit", isAuthed(GetMealEditHandler(logger, templateFiles, mealRepository)))
	mux.Handle("POST /meals/{id}", isAuthed(PutMealHandler(logger, mealService, mealRepository)))
	mux.Handle("POST /meals/{id}/delete", isAuthed(PostDeleteMealHandler(logger, mealRepository)))

	// Planner
	mux.Handle("GET /planner", isAuthed(GetPlannerHandler(logger, templateFiles, plannerService, userRepository)))
	mux.Handle("GET /planner/{date}", isAuthed(GetEditDayHandler(logger, templateFiles, plannerRepository, mealRepository)))
	mux.Handle("POST /planner/{date}", isAuthed(PostEditDayHandler(logger, plannerRepository, mealRepository)))
	mux.Handle("GET /planner/{date}/ingredients", isAuthed(GetPlannedIngredientsHandler(logger, plannerService)))

	// API
	mux.Handle("GET /api/ingredients", isAuthed(GetSearchIngredientsHandler(logger, mealRepository)))
	mux.Handle("GET /api/tags", isAuthed(GetSearchTagHandler(logger, mealRepository)))
	mux.Handle("GET /api/units", isAuthed(GetSearchUnitHandler(logger, mealRepository)))
}
