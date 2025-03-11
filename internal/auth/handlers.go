package auth

import (
	"encoding/json"
	"errors"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"github.com/jameswhoughton/meals/web/helpers"
)

type middleware func(http.Handler) http.Handler

func GetRegistrationHandler(templateFiles fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(templateFiles, "templates/layout.gohtml", "templates/form.gohtml", "templates/register.gohtml")

		if err != nil {
			w.Write([]byte("Template error: " + err.Error()))

			return
		}
		tmpl.ExecuteTemplate(w, "layout", nil)
	})
}

func PostRegistrationHandler(userService UserService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		form := UserFormCreate{
			Email:           r.FormValue("email"),
			Password:        r.FormValue("password"),
			PasswordConfirm: r.FormValue("passwordConfirm"),
			Name:            r.FormValue("name"),
		}

		_, err := userService.CreateUser(&form)

		if err != nil && errors.Is(err, ErrorFormInvalid{}) {
			errorJson, _ := json.Marshal(form.Errors)
			helpers.SetMessage(w, "errors", string(errorJson))

			http.Redirect(w, r, "/register", http.StatusFound)

			return
		}

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		helpers.SetMessage(w, "success", "Your account has been created, please login below")

		http.Redirect(w, r, "/login", http.StatusFound)
	})
}

func GetAccountHandler(templateFiles fs.FS, userService UserService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(templateFiles, "templates/layout.gohtml", "templates/navigation.gohtml", "templates/account.gohtml")

		if err != nil {
			w.Write([]byte("Template error: " + err.Error()))

			return
		}

		session, err := r.Cookie("session")

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		user, err := userService.GetUserFromSession(session.Value)

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		type templateData struct {
			Title   string
			Email   string
			Success string
			Errors  map[string]string
		}

		success, err := helpers.GetMessage(w, r, "success")

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		errorJson, err := helpers.GetMessage(w, r, "errors")

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		formErrors := map[string]string{}

		json.Unmarshal([]byte(errorJson), &formErrors)

		err = tmpl.ExecuteTemplate(w, "layout", templateData{
			Title:   "My Account",
			Email:   user.Email,
			Success: success,
			Errors:  formErrors,
		})

		if err != nil {
			log.Println(err)
		}
	})
}

func PutAccountHandler(userService UserService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := r.Cookie("session")

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		user, err := userService.GetUserFromSession(session.Value)

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		r.ParseForm()

		password := r.FormValue("password")
		passwordConfirm := r.FormValue("passwordConfirm")
		email := r.FormValue("email")
		name := r.FormValue("name")

		form := UserFormUpdate{
			Id:              user.Id,
			Password:        &password,
			PasswordConfirm: passwordConfirm,
			Email:           &email,
			Name:            &name,
		}

		err = userService.UpdateUser(&form)

		if err != nil && errors.Is(err, ErrorFormInvalid{}) {
			errorJson, _ := json.Marshal(form.Errors)
			helpers.SetMessage(w, "errors", string(errorJson))

			http.Redirect(w, r, "/account", http.StatusFound)

			return
		}

		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		helpers.SetMessage(w, "success", "you account has been updated")

		http.Redirect(w, r, "/account", http.StatusFound)
	})
}

func GetLoginHandler(templateFiles fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(templateFiles, "templates/layout.gohtml", "templates/form.gohtml", "templates/login.gohtml")

		if err != nil {
			w.Write([]byte("Template error: " + err.Error()))

			return
		}
		type loginData struct {
			Title   string
			Error   string
			Success string
		}

		errorMessage, err := helpers.GetMessage(w, r, "error")

		if err != nil {
			log.Println(err)

			errorMessage = "There was a problem with your request, please try again"
		}

		successMessage, err := helpers.GetMessage(w, r, "success")

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

func PostLoginHandler(userService UserService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		email, password := r.FormValue("email"), r.FormValue("password")

		user, err := userService.GetUserFromCredentials(email, password)

		if err != nil {
			helpers.SetMessage(w, "error", "credentials are invalid")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		session, err := userService.CreateSession(user.Id)

		if err != nil {
			w.WriteHeader(500)
			log.Print(err)
			return
		}

		userSession := http.Cookie{
			Name:     "session",
			Value:    session.SessionId,
			Path:     "/",
			MaxAge:   3600,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		}

		http.SetCookie(w, &userSession)

		http.Redirect(w, r, "/account", http.StatusFound)
	})
}

func GetLogoutHandler(userService UserService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentSesion, err := r.Cookie("session")

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		err = userService.sessionRepo.Destroy(currentSesion.Value)

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

func NewIsAuthenticatedMiddleware(userService UserService) middleware {
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

			_, err = userService.GetUserFromSession(session.Value)

			if err != nil {
				if errors.Is(err, ErrorSessionNotFound{}) {
					http.Redirect(w, r, "/login", http.StatusFound)
				}

				log.Println(err)
				http.Error(w, "server error", http.StatusInternalServerError)
			}

			next.ServeHTTP(w, r)
		})
	}
}
