package main

import (
	"context"
	"math-calc/internal/application"
	"os"
)

func main() {
	ctx := context.Background()
	os.Exit(mainWithExitCode(ctx))
}

func mainWithExitCode(ctx context.Context) int {
	app := application.NewApplication()
	return app.Run(ctx)
}
