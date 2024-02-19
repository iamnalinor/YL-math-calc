package main

import (
	"context"
	"math-calc/http/server"
	"math-calc/internal/application"
	"math-calc/internal/orchestrator"
	"os"
	"os/signal"
)

func main() {
	app := application.NewApplication()

	shutDownFunc, err := server.Run(app.Logger)
	if err != nil {
		app.Logger.Fatal(err.Error())
	}
	ctx, cancel := context.WithCancel(context.Background())
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Starting orchestrator and workers
	orc := orchestrator.New(app)
	go orc.Run()

	app.Logger.Println("Server started at localhost:8081")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	cancel()
	shutDownFunc(ctx)
	os.Exit(0)
}
