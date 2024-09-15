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
)

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

func NewRoutingEngine(db *gorm.DB, key *string) *gin.Engine {
	ginCore := gin.New()
	ginCore.RedirectTrailingSlash = false
	ginCore.RedirectFixedPath = false
	ginCore.Use(gin.Recovery())
	ginCore.Use(slogGin.New(slog.Default()))
	gzipMW := gzip.Gzip(gzip.DefaultCompression)

	ginCore.POST("/api/user/register", handlers.NewRegisterUserHandler(db, key))
	ginCore.POST("/api/user/login", handlers.NewLoginUserHandler(db, key))

	userGroup := ginCore.Group("/api/user")
	userGroup.Use(handlers.NewLoadUserByTokenMiddleware(db, key))
	userGroup.POST("/orders", handlers.NewAddAccrualOrderHandler(db))
	userGroup.GET("/orders", gzipMW, handlers.NewGetAccrualOrdersHandler(db))
	userGroup.GET("/balance", handlers.GetUserBalanceHandler)
	userGroup.POST("/balance/withdraw", handlers.NewAddWithdrawalOrderHandler(db))
	userGroup.GET("/withdrawals", gzipMW, handlers.NewGetWithdrawalOrdersHandler(db))

	return ginCore
}
