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
	userRepository       auth.UserRepository
	mealRepository       meals.MealRepository
	ingredientRepository meals.IngredientRepository
	tagRepository        meals.TagRepository
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	AddRoutes(mux, *s.userService, *s.mealService, s.userRepository, s.mealRepository, s.ingredientRepository, s.tagRepository)

	fmt.Printf("listening on port :%s\n", s.port)

	return http.ListenAndServe(":"+s.port, mux)
}

func NewServer(
	port string,
	userService *auth.UserService,
	mealService *meals.Service,
	userRepository auth.UserRepository,
	mealRepository meals.MealRepository,
	ingreidentRepository meals.IngredientRepository,
	tagRepository meals.TagRepository,
) *Server {
	return &Server{
		port:                 port,
		userService:          userService,
		mealService:          mealService,
		userRepository:       userRepository,
		mealRepository:       mealRepository,
		ingredientRepository: ingreidentRepository,
		tagRepository:        tagRepository,
	}
}
