package application

import (
	"context"
	"log"
	"math-calc/http/server"
	"math-calc/internal/config"
	"math-calc/internal/db"
	"math-calc/internal/orchestrator"
	"os"
	"os/signal"
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

func (a *Application) Run() int {
	// Starting web server

	shutDownFunc, err := server.Run(context.Background(), a.Logger)
	if err != nil {
		a.Logger.Fatal(err.Error())
		return 1
	}
	ctx, cancel := context.WithCancel(context.Background())
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Starting orchestrator and workers
	orc := orchestrator.New(a)
	go orc.Run()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	cancel()
	shutDownFunc(ctx)

	return 0
}

func setupLogger() *log.Logger {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	return logger
}
