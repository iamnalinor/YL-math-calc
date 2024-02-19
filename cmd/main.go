package main

import (
	"math-calc/internal/application"
	"os"
)

func main() {
	app := application.NewApplication()
	os.Exit(app.Run())
}
