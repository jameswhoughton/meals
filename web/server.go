package web

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/jameswhoughton/meals/internal/account"
	"github.com/jameswhoughton/meals/internal/meals"
	"github.com/jameswhoughton/meals/internal/planner"
)

func NewServer(
	ctx context.Context,
	port string,
	logger *slog.Logger,
	accountService *account.Service,
	mealService *meals.Service,
	sessionService *SessionService,
	plannerService *planner.Service,
	accountRepository account.Repository,
	mealRepository meals.Repository,
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
