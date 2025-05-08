package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/jameswhoughton/meals/database"
	"github.com/jameswhoughton/meals/internal/account"
	"github.com/jameswhoughton/meals/internal/meals"
	"github.com/jameswhoughton/meals/web"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	conn, err := sql.Open("sqlite3", "meals.db")
	defer conn.Close()

	if err != nil {
		log.Fatal(err)
	}

	err = database.Migrate(conn)

	if err != nil {
		log.Fatal(err)
	}

	// Signal context to detext an interupt (Ctrl + c) or terminate signal
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	accountRepository := database.NewAccountRespository(conn)
	sessionRepsoitory := database.NewSessionRepository(conn)
	mealRepository := database.NewMealRepository(conn)
	ingredientRepository := database.NewIngredientRepository(conn)
	tagRepository := database.NewTagRepository(conn)
	plannerRepository := database.NewPlannerRepository(conn)

	accountService := account.NewService(accountRepository)
	mealService := meals.NewService(mealRepository, ingredientRepository, tagRepository)
	sessionService := web.NewSessionService(accountRepository, sessionRepsoitory, 3600)

	ongoingCtx, stopOngoningGracefully := context.WithCancel(context.Background())
	port := "8000"
	server := web.NewServer(
		ongoingCtx,
		port,
		&accountService,
		&mealService,
		sessionService,
		accountRepository,
		mealRepository,
		ingredientRepository,
		tagRepository,
		sessionRepsoitory,
		plannerRepository,
	)

	go func() {
		log.Printf("Server starting on port :%s\n", port)

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
