package server

import (
	"errors"
	"github.com/fasdalf/train-go-musthave-diploma/internal/http/handlers"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	slogGin "github.com/samber/slog-gin"
	"gorm.io/gorm"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

func NewServer(db *gorm.DB, addr string, cryptoKey *string, exp time.Duration) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: NewRoutingEngine(db, cryptoKey, exp),
	}
}

func StartServer(srv *http.Server, wg *sync.WaitGroup) {
	wg.Add(1)
	go (func() {
		defer wg.Done()
		slog.Info("starting http server", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				slog.Info("server closed by interrupt signal")
			} else {
				slog.Error("server not started or stopped with error", "error", err)
				panic(err)
			}
		}
	})()
}

func NewRoutingEngine(db *gorm.DB, key *string, exp time.Duration) *gin.Engine {
	ginCore := gin.New()
	ginCore.RedirectTrailingSlash = false
	ginCore.RedirectFixedPath = false
	ginCore.Use(gin.Recovery())
	ginCore.Use(slogGin.New(slog.Default()))
	gzipMW := gzip.Gzip(gzip.DefaultCompression)

	ginCore.POST("/api/user/register", handlers.NewRegisterUser(db, key, exp))
	ginCore.POST("/api/user/login", handlers.NewLoginUser(db, key, exp))

	userGroup := ginCore.Group("/api/user")
	userGroup.Use(handlers.NewLoadUserByToken(db, key))
	userGroup.POST("/orders", handlers.NewAddAccrualOrder(db))
	userGroup.GET("/orders", gzipMW, handlers.NewGetAccrualOrders(db))
	userGroup.GET("/balance", handlers.GetUserBalance)
	userGroup.POST("/balance/withdraw", handlers.NewAddWithdrawalOrder(db))
	userGroup.GET("/withdrawals", gzipMW, handlers.NewGetWithdrawalOrders(db))

	return ginCore
}
