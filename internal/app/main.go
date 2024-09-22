package app

import (
	"context"
	"fmt"
	"github.com/fasdalf/train-go-musthave-diploma/internal/accrual"
	"github.com/fasdalf/train-go-musthave-diploma/internal/config"
	dbConn "github.com/fasdalf/train-go-musthave-diploma/internal/db/connection"
	"github.com/fasdalf/train-go-musthave-diploma/internal/db/repository"
	accrualClient "github.com/fasdalf/train-go-musthave-diploma/internal/http/accrual"
	"github.com/fasdalf/train-go-musthave-diploma/internal/http/server"
	"gorm.io/gorm"
	"net/http"
	"sync"
)

type App struct {
	ctx             context.Context
	wg              *sync.WaitGroup
	cfg             *config.Config
	db              *gorm.DB
	orderRepository accrual.OrderRepository
	accrualClient   accrual.AccrualClient
	httpServer      *http.Server
}

func NewApp(ctx context.Context) (*App, error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	db, err := dbConn.NewConnection(ctx, cfg.DatabaseURI)
	if err != nil {
		return nil, fmt.Errorf("failed to init DB connection: %w", err)
	}
	or := repository.NewOrderRepository(db)
	ac := accrualClient.NewAccrualClient(&cfg.Accrual.URL)

	srv := server.NewServer(db, cfg.Addr, &cfg.CryptoKey, cfg.TokenExp)

	return &App{
		ctx:             ctx,
		cfg:             cfg,
		db:              db,
		orderRepository: or,
		accrualClient:   ac,
		wg:              &sync.WaitGroup{},
		httpServer:      srv,
	}, nil
}

func (a *App) Run() {
	server.StartServer(a.httpServer, a.wg)
	accrual.StartChecker(a.ctx, a.wg, a.accrualClient, a.orderRepository, &a.cfg.Accrual)
}

func (a *App) Shutdown(cancel context.CancelFunc) error {
	err := a.httpServer.Shutdown(a.ctx)
	if err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}
	cancel()
	// Let worker's transactions to end gracefully.
	a.wg.Wait()
	return nil
}
