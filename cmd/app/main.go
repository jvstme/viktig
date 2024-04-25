package main

import (
	"fmt"
	"log/slog"
	"viktig/internal/app"
	_ "viktig/internal/logger"
)

func main() {
	a := app.New()
	err := a.Run()
	if err != nil {
		slog.Error(fmt.Sprintf("error running app: %+v", err))
	}
}
