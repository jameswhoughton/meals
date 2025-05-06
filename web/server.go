package web

import (
	"context"
	"net"
	"net/http"

	"github.com/jameswhoughton/meals/internal/auth"
	"github.com/jameswhoughton/meals/internal/meals"
)

func NewServer(
	ctx context.Context,
	port string,
	userService *auth.UserService,
	mealService *meals.Service,
	userRepository auth.UserRepository,
	mealRepository meals.MealRepository,
	ingredientRepository meals.IngredientRepository,
	tagRepository meals.TagRepository,
) *http.Server {
	mux := http.NewServeMux()

	AddRoutes(mux, *userService, *mealService, userRepository, mealRepository, ingredientRepository, tagRepository)

	return &http.Server{
		Addr:        ":" + port,
		Handler:     mux,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}
}
