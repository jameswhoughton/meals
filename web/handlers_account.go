package web

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strconv"

	"github.com/jameswhoughton/meals/internal/account"
)

type middleware func(http.Handler) http.Handler

func GetRegistrationHandler(templateFiles fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(
			templateFiles,
			"templates/layout_guest.gohtml",
			"templates/pages/auth/partials/form.gohtml",
			"templates/pages/auth/register.gohtml",
		)

		if err != nil {
			w.Write([]byte("Template error: " + err.Error()))

			return
		}

		formJson, err := getMessage(w, r, "formData")

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		formData := account.UserFormCreate{}

		if formJson != "" {
			json.Unmarshal([]byte(formJson), &formData)
		}

		type templateData struct {
			Title string
			Form  account.UserFormCreate
		}

		tmpl.ExecuteTemplate(w, "layout", templateData{
			Title: "Register",
			Form:  formData,
		})
	})
}

func PostRegistrationHandler(service account.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		form := account.UserFormCreate{
			Email:           r.FormValue("email"),
			Password:        r.FormValue("password"),
			PasswordConfirm: r.FormValue("passwordConfirm"),
			Name:            r.FormValue("name"),
		}

		_, err := service.CreateUser(r.Context(), &form)

		if err != nil && errors.Is(err, account.ErrUserFormInvalid) {
			formJson, _ := json.Marshal(form)
			setMessage(w, "formData", string(formJson))

			http.Redirect(w, r, "/register", http.StatusFound)

			return
		}

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		setMessage(w, "success", "Your account has been created, please login below")

		http.Redirect(w, r, "/login", http.StatusFound)
	})
}

func GetAccountHandler(templateFiles fs.FS, service SessionService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(templateFiles, "templates/layout.gohtml", "templates/navigation.gohtml", "templates/pages/auth/account.gohtml")

		if err != nil {
			w.Write([]byte("Template error: " + err.Error()))

			return
		}

		session, err := r.Cookie("session")

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		user, err := service.UserFromSession(r.Context(), session.Value)

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		type templateData struct {
			Title   string
			Success string
			Form    account.UserFormUpdate
		}

		success, err := getMessage(w, r, "success")

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		formJson, err := getMessage(w, r, "formData")

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		formData := account.UserFormUpdate{
			Name:         user.Name,
			Email:        user.Email,
			MealStartDay: user.MealStartDay,
		}

		if formJson != "" {
			json.Unmarshal([]byte(formJson), &formData)
		}

		err = tmpl.ExecuteTemplate(w, "layout", templateData{
			Title:   "My Account",
			Form:    formData,
			Success: success,
		})

		if err != nil {
			log.Println(err)
		}
	})
}

func PutAccountHandler(accountService account.Service, sessionService SessionService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := r.Cookie("session")

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		user, err := sessionService.UserFromSession(r.Context(), session.Value)

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		r.ParseForm()

		password := r.FormValue("password")
		passwordConfirm := r.FormValue("passwordConfirm")

		mealStartDay, err := strconv.Atoi(r.FormValue("mealStartDay"))

		if err != nil {
			http.Error(w, "mealStartDay must be an integer", http.StatusBadRequest)

			return
		}

		form := account.UserFormUpdate{
			Id:           user.Id,
			Email:        r.FormValue("email"),
			Name:         r.FormValue("name"),
			MealStartDay: mealStartDay,
		}

		if password != "" {
			form.Password = &password
			form.PasswordConfirm = passwordConfirm
		}

		err = accountService.UpdateUser(r.Context(), &form)

		if err != nil && errors.Is(err, account.ErrUserFormInvalid) {
			formJson, _ := json.Marshal(form)
			setMessage(w, "formData", string(formJson))

			http.Redirect(w, r, "/account", http.StatusFound)

			return
		}

		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		setMessage(w, "success", "you account has been updated")

		http.Redirect(w, r, "/account", http.StatusFound)
	})
}

func GetLoginHandler(templateFiles fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(
			templateFiles,
			"templates/layout_guest.gohtml",
			"templates/pages/auth/partials/form.gohtml",
			"templates/pages/auth/login.gohtml",
		)

		if err != nil {
			w.Write([]byte("Template error: " + err.Error()))

			return
		}
		type loginData struct {
			Title   string
			Error   string
			Success string
		}

		errorMessage, err := getMessage(w, r, "error")

		if err != nil {
			log.Println(err)

			errorMessage = "There was a problem with your request, please try again"
		}

		successMessage, err := getMessage(w, r, "success")

		if err != nil {
			log.Println(err)

			errorMessage = "There was a problem with your request, please try again"
		}

		err = tmpl.ExecuteTemplate(w, "layout", loginData{
			Title:   "Meals - Login",
			Error:   errorMessage,
			Success: successMessage,
		})

		if err != nil {
			log.Print(err)
		}
	})
}

func PostLoginHandler(accountService account.Service, sessionService SessionService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		email, password := r.FormValue("email"), r.FormValue("password")

		user, err := accountService.GetUserFromCredentials(r.Context(), email, password)

		if err != nil {
			log.Println(err)
			setMessage(w, "error", "credentials are invalid")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		session, err := sessionService.CreateForUser(r.Context(), user.Id)

		if err != nil {
			w.WriteHeader(500)
			log.Print(err)
			return
		}

		sessionCookie := http.Cookie{
			Name:     "session",
			Value:    session.SessionId,
			Path:     "/",
			MaxAge:   3600,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		}

		http.SetCookie(w, &sessionCookie)

		http.Redirect(w, r, "/planner", http.StatusFound)
	})
}

func GetLogoutHandler(service account.Service, session SessionRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentSesion, err := r.Cookie("session")

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		err = session.Destroy(r.Context(), currentSesion.Value)

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		userSession := http.Cookie{
			Name:     "session",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		}

		http.SetCookie(w, &userSession)

		http.Redirect(w, r, "/login", http.StatusFound)
	})
}

func NewIsAuthenticatedMiddleware(accountService account.Service, sessionService SessionService) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			session, err := r.Cookie("session")

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

			user, err := sessionService.UserFromSession(r.Context(), session.Value)

			if err != nil {
				if errors.Is(err, ErrSessionNotFound) {
					setMessage(w, "error", "Session has expired, please login again")
					http.Redirect(w, r, "/login", http.StatusFound)

					return
				}

				log.Println(err)
				http.Error(w, "server error", http.StatusInternalServerError)

				return
			}

			// Refresh the token cookie
			userSession := http.Cookie{
				Name:     "session",
				Value:    session.Value,
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
			}

			http.SetCookie(w, &userSession)

			ctx := context.WithValue(r.Context(), "userId", user.Id)

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
