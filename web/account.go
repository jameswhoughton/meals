package web

import (
	"encoding/json"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	meals "github.com/jameswhoughton/meals"
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

		form := meals.UserForm{
			Email:           r.FormValue("email"),
			Password:        r.FormValue("password"),
			PasswordConfirm: r.FormValue("passwordConfirm"),
			Name:            r.FormValue("name"),
		}

		if !form.Validate(userService) {
			errorJson, _ := json.Marshal(form.Errors)
			setMessage(w, "errors", string(errorJson))

			http.Redirect(w, r, "/register", http.StatusFound)

			return

		}

		_, err := userService.Add(form)

		if err != nil {
			log.Fatal(err)
		}

		setMessage(w, "success", "Your account has been created, please login below")

		http.Redirect(w, r, "/login", http.StatusFound)
	})
}

func getAccountHandler(templateFiles fs.FS, sessionService meals.SessionService) http.Handler {
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

		user, err := sessionService.GetUser(session.Value)

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

func putAccountHandler(userService meals.UserService, sessionService meals.SessionService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := r.Cookie("session")

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		user, err := sessionService.GetUser(session.Value)

		if err != nil {
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		r.ParseForm()

		form := meals.UserForm{
			Password:        r.FormValue("password"),
			PasswordConfirm: r.FormValue("passwordConfirm"),
			Email:           r.FormValue("email"),
			Name:            r.FormValue("name"),
		}

		if !form.Validate(user, userService) {
			errorJson, _ := json.Marshal(form.Errors)
			setMessage(w, "errors", string(errorJson))

			http.Redirect(w, r, "/account", http.StatusFound)

			return
		}

		err = userService.Update(user.Id, form)

		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
		}

		setMessage(w, "success", "you account has been updated")

		http.Redirect(w, r, "/account", http.StatusFound)
	})
}
