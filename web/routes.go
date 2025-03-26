package web

import (
	"embed"
	"errors"
	"log"
	"net/http"

	"github.com/jameswhoughton/meals/internal/auth"
	"github.com/jameswhoughton/meals/internal/meals"
)

//go:embed templates/*.gohtml
//go:embed templates/**/**/*.gohtml
//go:embed templates/**/**/**/*.gohtml
var templateFiles embed.FS

//go:embed static/*
var publicFiles embed.FS

func AddRoutes(
	mux *http.ServeMux,
	userService auth.UserService,
	mealService meals.Service,
) {
	// Middleware
	isAuthed := auth.NewIsAuthenticatedMiddleware(userService)

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
	mux.Handle("GET /login", auth.GetLoginHandler(templateFiles))
	mux.Handle("POST /login", auth.PostLoginHandler(userService))
	mux.Handle("GET /register", auth.GetRegistrationHandler(templateFiles))
	mux.Handle("POST /register", auth.PostRegistrationHandler(userService))
	mux.Handle("GET /logout", auth.GetLogoutHandler(userService))

	// Account
	mux.Handle("GET /account", isAuthed(auth.GetAccountHandler(templateFiles, userService)))
	mux.Handle("POST /account", isAuthed(auth.PutAccountHandler(userService)))

	// Meals
	mux.Handle("GET /meals/create", isAuthed(meals.GetCreateMealHandler(templateFiles)))
	mux.Handle("POST /meals/create", isAuthed(meals.PostMealHandler(mealService)))
}
