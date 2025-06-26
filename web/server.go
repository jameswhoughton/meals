package web

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/jameswhoughton/meals"
	"github.com/jameswhoughton/meals/internal/planner"
)

func NewServer(
	ctx context.Context,
	port string,
	logger *slog.Logger,
	accountService *meals.UserService,
	mealService *meals.MealService,
	sessionService *SessionService,
	plannerService *planner.Service,
	accountRepository meals.UserRepository,
	mealRepository meals.MealRepository,
	sessionRepository SessionRepository,
	plannerRepository planner.Repository,
) *http.Server {
	mux := http.NewServeMux()

	AddRoutes(mux, logger, *accountService, *mealService, *sessionService, *plannerService, accountRepository, mealRepository, sessionRepository, plannerRepository)

	return &http.Server{
		Addr:              ":" + port,
		ReadTimeout:       500 * time.Millisecond,
		ReadHeaderTimeout: 500 * time.Millisecond,
		Handler:           http.TimeoutHandler(mux, 5*time.Second, "timeout"),
		BaseContext:       func(_ net.Listener) context.Context { return ctx },
	}
}
