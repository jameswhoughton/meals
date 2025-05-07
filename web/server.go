package web

import (
	"context"
	"net"
	"net/http"

	"github.com/jameswhoughton/meals/internal/account"
	"github.com/jameswhoughton/meals/internal/meals"
	"github.com/jameswhoughton/meals/internal/planner"
)

func NewServer(
	ctx context.Context,
	port string,
	accountService *account.Service,
	mealService *meals.Service,
	sessionService *SessionService,
	accountRepository account.Repository,
	mealRepository meals.MealRepository,
	ingredientRepository meals.IngredientRepository,
	tagRepository meals.TagRepository,
	sessionRepository SessionRepository,
	plannerRepository planner.Repository,
) *http.Server {
	mux := http.NewServeMux()

	AddRoutes(mux, *accountService, *mealService, *sessionService, accountRepository, mealRepository, ingredientRepository, tagRepository, sessionRepository, plannerRepository)

	return &http.Server{
		Addr:        ":" + port,
		Handler:     mux,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}
}
