package main

import (
	"context"
	"github.com/fasdalf/train-go-musthave-diploma/internal/accrual"
	"github.com/fasdalf/train-go-musthave-diploma/internal/config"
	dbConn "github.com/fasdalf/train-go-musthave-diploma/internal/db/connection"
	"github.com/fasdalf/train-go-musthave-diploma/internal/http/server"
	"log/slog"
	"os"
	"os/signal"
	"sync"
)

func main() {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()
	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("failed to parse config", "error", err)
		panic(err)
	}
	slog.Info("init DB connection")
	db, err := dbConn.NewConnection(ctx, cfg.DatabaseURI)
	if err != nil {
		slog.Error("failed to init DB", "error", err)
		panic(err)
	}
	slog.Info("DB successfully migrated")

	wg := &sync.WaitGroup{}

	slog.Debug("initializing http router")
	srv := server.NewServer(db, cfg.Addr, &cfg.CryptoKey, cfg.TokenExp)

	server.StartServer(srv, wg)
	accrual.StartChecker(ctx, wg, db, &cfg.Accrual)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	slog.Info("interrupt signal received")
	signal.Stop(quit)
	slog.Info("shutting down http server...")
	err = srv.Shutdown(ctx)
	if err != nil {
		slog.Error("Server shutdown error:", "error", err)
	}
	ctxCancel()
	slog.Info("waiting for bg processes...")
	// Let worker's transactions to end gracefully.
	wg.Wait()
}
