package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luponetn/paycore/internal/auth"
	"github.com/luponetn/paycore/internal/config"
	"github.com/luponetn/paycore/internal/db"
	"github.com/luponetn/paycore/internal/store"
	"github.com/luponetn/paycore/internal/tasks"
	"github.com/luponetn/paycore/internal/transfer"
	"github.com/luponetn/paycore/internal/wallet"
)

type Application struct {
	config *config.Config
	db     *pgxpool.Pool
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	slog.Info("config loaded successfully", "config", cfg)

	//setup database connection
	dbConn, err := db.ConnDb(cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbConn.Close()

	slog.Info("database connection established successfully")

	app := Application{
		config: cfg,
		db:     dbConn,
	}

	router := app.SetupRouter()
	queries := db.New(dbConn)

	//postgresStore setup for main app db calls
	postgresStore := store.NewPostgresStore(dbConn, queries)

	//setup task client
	taskClient := tasks.NewTaskClient(cfg.RedisAddr)
	defer taskClient.Close()

	//register service
	authSvc := auth.NewService(postgresStore, taskClient, cfg)
	transferSvc := transfer.NewService(postgresStore)
	walletSvc := wallet.NewService(postgresStore)

	//register handler
	authHandler := auth.NewHandler(authSvc)
	transferHandler := transfer.NewHandler(transferSvc)
	walletHandler := wallet.NewHandler(walletSvc)

	//register routes
	auth.RegisterRoutes(router, authHandler)
	transfer.RegisterRoutes(router, transferHandler, cfg.JWTAccessSecret)
	wallet.RegisterRoutes(router, walletHandler, cfg.JWTAccessSecret)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		slog.Info("Starting server on port: " + cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to listen and serve", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown:", "error", err)
	}

	slog.Info("Server exiting")
}
