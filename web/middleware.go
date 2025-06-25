package web

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/jameswhoughton/meals/internal/account"
)

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

			ctx := context.WithValue(r.Context(), "user", user)

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

func UserFromContext(ctx context.Context) *account.User {
	user, ok := ctx.Value("user").(account.User)

	if !ok {
		return nil
	}

	return &user
}
