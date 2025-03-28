package web

import (
	"fmt"
	"net/http"

	"github.com/jameswhoughton/meals/internal/auth"
	"github.com/jameswhoughton/meals/internal/meals"
)

type Server struct {
	port                 string
	userService          *auth.UserService
	mealService          *meals.Service
	mealRepository       meals.MealRepository
	ingredientRepository meals.IngredientRepository
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	AddRoutes(mux, *s.userService, *s.mealService, s.mealRepository, s.ingredientRepository)

	fmt.Printf("listening on port :%s\n", s.port)

	return http.ListenAndServe(":"+s.port, mux)
}

func NewServer(
	port string,
	userService *auth.UserService,
	mealService *meals.Service,
	mealRepository meals.MealRepository,
	ingreidentRepository meals.IngredientRepository,
) *Server {
	return &Server{
		port:                 port,
		userService:          userService,
		mealService:          mealService,
		mealRepository:       mealRepository,
		ingredientRepository: ingreidentRepository,
	}
}
