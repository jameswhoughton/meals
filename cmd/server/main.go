package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jameswhoughton/meals/database"
	"github.com/jameswhoughton/meals/internal/account"
	"github.com/jameswhoughton/meals/internal/meals"
	"github.com/jameswhoughton/meals/internal/planner"
	"github.com/jameswhoughton/meals/web"
)

type config struct {
	port string
	dsn  string
}

// Retrieve environment configuration.
//
// Either returns valid config or will panic.
func getConfig() config {
	port := os.Getenv("MEALS_PORT")

	if port == "" {
		panic("MEALS_PORT environment variable is missing or blank")
	}

	dsn := os.Getenv("MEALS_DSN")

	if dsn == "" {
		panic("MEALS_DSN environment variable is missing or blank")
	}

	return config{
		port: port,
		dsn:  dsn,
	}
}

func main() {
	config := getConfig()

	conn, err := sql.Open("mysql", config.dsn)
	defer conn.Close()

	if err != nil {
		panic(err)
	}

	err = database.Migrate(conn)

	if err != nil {
		panic(err)
	}

	// Signal context to detext an interupt (Ctrl + c) or terminate signal
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	accountRepository := database.NewAccountRespository(conn)
	sessionRepsoitory := database.NewSessionRepository(conn)
	mealRepository := database.NewMealRepository(conn)
	plannerRepository := database.NewPlannerRepository(conn)

	accountService := account.NewService(accountRepository)
	mealService := meals.NewService(mealRepository)
	sessionService := web.NewSessionService(accountRepository, sessionRepsoitory, 3600)
	plannerSerivce := planner.NewService(plannerRepository)

	ongoingCtx, stopOngoningGracefully := context.WithCancel(context.Background())
	server := web.NewServer(
		ongoingCtx,
		config.port,
		&accountService,
		&mealService,
		sessionService,
		plannerSerivce,
		accountRepository,
		mealRepository,
		sessionRepsoitory,
		plannerRepository,
	)

	go func() {
		log.Printf("Server starting on port :%s\n", config.port)

		// Ignore ErrServerClosed as this happens when the server is expectedly shutdown
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error running server: %v", err)
		}
	}()

	<-rootCtx.Done()
	// Need to call stop here as well to unregister the signal so that a second SIGINT will behave
	// in the default manner and kiil the program.
	stop()

	log.Println("Gracefully shutting down")

	// Create a timeout context allowing 10s for the server to stop receiving requests and
	// finish processing existing requests.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = server.Shutdown(shutdownCtx)

	// Inform any in progress requests to stop processing
	stopOngoningGracefully()

	if err != nil {
		log.Println("Requests have not finished")
	}

	log.Println("Server shut down gracefully")
}
