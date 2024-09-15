package main

import (
	"context"
	"github.com/fasdalf/train-go-musthave-diploma/internal/accrual"
	"github.com/fasdalf/train-go-musthave-diploma/internal/config"
	dbConn "github.com/fasdalf/train-go-musthave-diploma/internal/db/connection"
	"github.com/fasdalf/train-go-musthave-diploma/internal/http/server"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
)

func main() {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()
	cfg := config.GetConfig()
	slog.Info("init DB connection")
	db, err := dbConn.NewConnection(ctx, cfg.DatabaseURI)
	if err != nil {
		slog.Error("failed to init DB", "error", err)
		panic(err)
	}
	slog.Info("DB successfully migrated")

	wg := &sync.WaitGroup{}

	slog.Debug("initializing http router")
	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: server.NewRoutingEngine(db, &cfg.CryptoKey),
	}

	server.StartServer(srv, wg)
	accrual.StartChecker(ctx, wg, db, cfg.RemoteAccrualAddr)

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
	// it could be errgroup.Group.Wait but I don't want to collect errors
	wg.Wait()
}
