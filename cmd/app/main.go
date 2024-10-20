package main

import (
	"fmt"
	"log/slog"
	"os"
	"viktig/internal/app"
	_ "viktig/internal/logger"
)

func main() {
	a, err := app.New()
	if err != nil {
		slog.Error(fmt.Sprintf("error running app: %+v", err))
		os.Exit(1)
	} else {
		a.Run()
	}
}
