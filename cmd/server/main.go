package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jameswhoughton/meals/database"
	"github.com/jameswhoughton/meals/internal"
	"github.com/jameswhoughton/meals/internal/account"
	"github.com/jameswhoughton/meals/internal/meals"
	"github.com/jameswhoughton/meals/internal/planner"
	"github.com/jameswhoughton/meals/web"
	"github.com/joho/godotenv"
)

type config struct {
	port string
	dsn  string
}

// Retrieve environment configuration.
//
// Either returns valid config or will panic.
func getConfig() config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	port := os.Getenv("APP_PORT")

	if port == "" {
		panic("APP_PORT environment variable is missing or blank")
	}

	dsn := os.Getenv("DB_USERNAME")

	if os.Getenv("DB_PASSWORD") != "" {
		dsn += ":" + os.Getenv("DB_PASSWORD")
	}

	dsn = fmt.Sprintf(
		"%s@tcp(%s:%s)/meals?parseTime=true",
		dsn,
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
	)

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

	// Ping the database to ensure it is ready to receive queries.
	for i := range 10 {
		err = conn.Ping()

		if err == nil {
			break
		}

		fmt.Printf("Waiting for DB... attempt %d\n", i)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		panic("Database inactive: " + err.Error())
	}

	err = database.Migrate(conn)

	if err != nil {
		panic(err)
	}

	// Configure the logger
	logFile, err := os.OpenFile("application.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)

	if err != nil {
		panic(err)
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			log.Printf("unable to close log file: %v", err)
		}
	}()

	logger := internal.NewApplicationLogger(logFile)

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
		logger,
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
			logger.LogAttrs(
				context.TODO(),
				slog.LevelError,
				"error starting server",
				slog.Any("err", err),
			)
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
