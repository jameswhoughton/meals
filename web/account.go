package web

import (
	"encoding/json"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	meals "github.com/jameswhoughton/meals"
	"golang.org/x/crypto/bcrypt"
)

func getRegistrationHandler(templateFiles fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(templateFiles, "templates/layout.gohtml", "templates/form.gohtml", "templates/register.gohtml")

		if err != nil {
			w.Write([]byte("Template error: " + err.Error()))

			return
		}
		tmpl.ExecuteTemplate(w, "layout", nil)
	})
}

func postRegistrationHandler(userService meals.UserService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		hash, err := bcrypt.GenerateFromPassword([]byte(r.FormValue("password")), bcrypt.MinCost)
		if err != nil {
			log.Println(err)
		}

		user := meals.User{
			Email:    r.FormValue("email"),
			Password: string(hash),
		}

		_, err = userService.Add(user)

		if err != nil {
			log.Fatal(err)
		}

		setMessage(w, "success", "Your account has been created, please login below")

		http.Redirect(w, r, "/login", http.StatusFound)
	})
}

func getAccountHandler(templateFiles fs.FS, userService meals.UserService) http.Handler {
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

		user, err := userService.GetFromSessionId(session.Value)

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		type templateData struct {
			Title   string
			Email   string
			Success string
			Errors  formErrors
		}

		success, err := getMessage(w, r, "success")

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		errorJson, err := getMessage(w, r, "errors")

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		formErrors := formErrors{}

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

type formErrors struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type userForm struct {
	password        string
	passwordConfirm string
	email           string
	errors          formErrors
}

func (uf *userForm) isValid(currentUser meals.User, userService meals.UserService) bool {
	if uf.password != uf.passwordConfirm {
		uf.errors.Password = "password and confirm do not match"

		return false
	}

	if currentUser.Email == uf.email {
		return true
	}

	existingUser, _ := userService.GetFromEmail(uf.email)

	if existingUser.Id > 0 {
		uf.errors.Email = "email already in use"

		return false
	}

	return true
}

func putAccountHandler(userService meals.UserService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := r.Cookie("session")

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		user, err := userService.GetFromSessionId(session.Value)

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		r.ParseForm()

		form := userForm{
			password:        r.FormValue("password"),
			passwordConfirm: r.FormValue("passwordConfirm"),
			email:           r.FormValue("email"),
		}

		if !form.isValid(user, userService) {
			errorJson, _ := json.Marshal(form.errors)
			setMessage(w, "errors", string(errorJson))

			http.Redirect(w, r, "/account", http.StatusFound)

			return
		}

		if form.password != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(form.password), bcrypt.MinCost)
			if err != nil {
				log.Println(err)
			}

			err = userService.UpdatePassword(user, string(hash))

			if err != nil {
				log.Println(err)
				http.Error(w, "server error", http.StatusInternalServerError)
			}
		}

		if user.Email != form.email {
			err = userService.UpdateEmail(user, form.email)

			if err != nil {
				log.Println(err)
				http.Error(w, "server error", http.StatusInternalServerError)
			}
		}

		setMessage(w, "success", "you account has been updated")

		http.Redirect(w, r, "/account", http.StatusFound)
	})
}
