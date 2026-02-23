package main

import (
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luponetn/paycore/internal/config"
	"github.com/luponetn/paycore/internal/db"
)

type Application struct {
	logger *slog.Logger
	config *config.Config
	db     *pgxpool.Pool
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Info("config loaded successfully", "config", cfg)

	//setup database connection
	dbConn, err := db.ConnDb(cfg, logger)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbConn.Close()

	logger.Info("database connection established successfully")

	app := Application{
		logger: logger,
		config: cfg,
		db:     dbConn,
	}

	router := app.SetupRouter()

	logger.Info("Starting server on port:", "port", cfg.Port)

	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Error("The server was unable to start up", "error", err)
		os.Exit(1)
	}
}
