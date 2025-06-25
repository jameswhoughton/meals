package web

import (
	"encoding/json"
	"errors"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jameswhoughton/meals/internal/account"
)

type middleware func(http.Handler) http.Handler

func GetRegistrationHandler(logger *slog.Logger, templateFiles fs.FS) http.Handler {
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
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"unable to get flash message",
				slog.Any("err", err),
			)
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

func PostRegistrationHandler(logger *slog.Logger, service account.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		form := account.UserFormCreate{
			Email:           r.FormValue("email"),
			Password:        r.FormValue("password"),
			PasswordConfirm: r.FormValue("passwordConfirm"),
			Name:            r.FormValue("name"),
		}

		user, err := service.CreateUser(r.Context(), &form)

		if err != nil && errors.Is(err, account.ErrUserFormInvalid) {
			formJson, _ := json.Marshal(form)
			setMessage(w, "formData", string(formJson))

			http.Redirect(w, r, "/register", http.StatusFound)

			return
		}

		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"unable to create user",
				slog.Any("err", err),
			)
			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		logger.LogAttrs(
			r.Context(),
			slog.LevelInfo,
			"New user created",
			slog.Int("userId", user.Id),
		)

		setMessage(w, "success", "Your account has been created, please login below")

		http.Redirect(w, r, "/login", http.StatusFound)
	})
}

func GetAccountHandler(logger *slog.Logger, templateFiles fs.FS, service SessionService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(templateFiles, "templates/layout.gohtml", "templates/navigation.gohtml", "templates/pages/auth/account.gohtml")

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

		type templateData struct {
			Title   string
			Success string
			Form    account.UserFormUpdate
		}

		success, err := getMessage(w, r, "success")

		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"unable to get success message",
				slog.Any("err", err),
			)

			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		formJson, err := getMessage(w, r, "formData")

		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"unable to get form data",
				slog.Any("err", err),
			)

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

		data := templateData{
			Title:   "My Account",
			Form:    formData,
			Success: success,
		}

		err = tmpl.ExecuteTemplate(w, "layout", data)

		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"error executing template",
				slog.Any("err", err),
				slog.Any("templateData", data),
			)
		}
	})
}

func PutAccountHandler(logger *slog.Logger, accountService account.Service, sessionService SessionService) http.Handler {
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
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"error updating user",
				slog.Any("err", err),
				slog.Int("userId", form.Id),
			)

			http.Error(w, "server error", http.StatusInternalServerError)
		}

		setMessage(w, "success", "you account has been updated")

		http.Redirect(w, r, "/account", http.StatusFound)
	})
}

func GetLoginHandler(logger *slog.Logger, templateFiles fs.FS) http.Handler {
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
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"unable to get error message",
				slog.Any("err", err),
			)

			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		successMessage, err := getMessage(w, r, "success")

		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"unable to get success message",
				slog.Any("err", err),
			)

			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		data := loginData{
			Title:   "Meals - Login",
			Error:   errorMessage,
			Success: successMessage,
		}

		err = tmpl.ExecuteTemplate(w, "layout", data)

		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"error executing template",
				slog.Any("err", err),
				slog.Any("templateData", data),
			)
		}
	})
}

func PostLoginHandler(logger *slog.Logger, accountService account.Service, sessionService SessionService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		email, password := r.FormValue("email"), r.FormValue("password")

		user, err := accountService.GetUserFromCredentials(r.Context(), email, password)

		if err != nil {
			setMessage(w, "error", "credentials are invalid")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		session, err := sessionService.CreateForUser(r.Context(), user.Id)

		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"unable to create session for user",
				slog.Any("err", err),
				slog.Int("userId", user.Id),
			)

			http.Error(w, "server error", http.StatusInternalServerError)
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

func GetLogoutHandler(logger *slog.Logger, service account.Service, session SessionRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentSesion, err := r.Cookie("session")

		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"unable to fetch session cookie",
				slog.Any("err", err),
			)

			http.Error(w, "server error", http.StatusInternalServerError)

			return
		}

		err = session.Destroy(r.Context(), currentSesion.Value)

		if err != nil {
			user := UserFromContext(r.Context())

			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"unable to destroy session for user",
				slog.Any("err", err),
				slog.Int("userId", user.Id),
			)

			http.Error(w, "server error", http.StatusInternalServerError)

			return
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
