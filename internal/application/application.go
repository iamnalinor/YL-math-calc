package application

import (
	"log"
	"math-calc/internal/config"
	"math-calc/internal/db"
	"os"
)

type Application struct {
	Config   config.Config
	Logger   *log.Logger
	Database *db.Database
}

func NewApplication() *Application {
	logger := setupLogger()

	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		logger.Fatal(err)
	}

	database, err := db.New()
	if err != nil {
		logger.Fatal(err)
	}

	return &Application{
		Config:   cfg,
		Logger:   logger,
		Database: database,
	}
}

func setupLogger() *log.Logger {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	return logger
}
