package main

import (
	"context"
	"github.com/fasdalf/train-go-musthave-diploma/internal/app"
	"log/slog"
	"os"
	"os/signal"
)

func main() {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()
	mainApp, err := app.NewApp(ctx)
	if err != nil {
		slog.Error("failed to init app", "error", err)
		panic(err)
	}
	mainApp.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	slog.Info("interrupt signal received")
	err = mainApp.Shutdown(ctxCancel)
	if err != nil {
		slog.Error("failed to init app", "error", err)
	}
}
