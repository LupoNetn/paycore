package main

import (
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luponetn/paycore/internal/auth"
	"github.com/luponetn/paycore/internal/config"
	"github.com/luponetn/paycore/internal/db"
	"github.com/luponetn/paycore/internal/tasks"
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

	//setup task client
	taskClient := tasks.NewTaskClient(cfg.RedisAddr)
	defer taskClient.Close()

	//register service
	authSvc := auth.NewService(queries, cfg, taskClient)

	//register handler
	authHandler := auth.NewHandler(authSvc)

	//register routes
	auth.RegisterRoutes(router, authHandler)

	slog.Info("Starting server on port: " + cfg.Port)

	if err := router.Run(":" + cfg.Port); err != nil {
		slog.Error("The server was unable to start up", "error", err)
		os.Exit(1)
	}
}
