package handlers

import (
	"bufio"
	"errors"
	"github.com/fasdalf/train-go-musthave-diploma/internal/db/entity"
	"github.com/fasdalf/train-go-musthave-diploma/pkg/luhn"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

// NewAddAccrualOrderHandler gin handler
func NewAddAccrualOrderHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		b := bufio.NewReader(c.Request.Body)
		s, err := b.ReadString(0)
		if err != nil && !errors.Is(err, io.EOF) {
			slog.Error("read error", "error", err)
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("read error"))
			return
		}
		number := strings.TrimSpace(s)
		if len(number) == 0 {
			slog.Error("number is whitespace")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("number is whitespace"))
			return
		}
		if !luhn.LuhnAlgorithm(number) {
			slog.Error("number is mystyped")
			_ = c.AbortWithError(http.StatusUnprocessableEntity, errors.New("number is mystyped"))
			return
		}

		order := &entity.Order{OrderNumber: number}
		if err := db.Where(order).Find(order).Error; err != nil {
			slog.Error("failed to find order", "err", err)
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("failed to find order"))
			return
		}

		user, err := getUserFromContext(c)
		if err != nil {
			slog.Error("user in context is not a user type", "error", err)
			_ = c.AbortWithError(http.StatusInternalServerError, errors.New("user in context is not a user type"))
			return
		}

		if order.ID != 0 {
			if order.UserID == user.ID {
				c.JSON(http.StatusOK, gin.H{"status": order.FetchStatus})
				return
			}
			slog.Error("order already exists")
			_ = c.AbortWithError(http.StatusConflict, errors.New("order already exists"))
			return
		}

		order = &entity.Order{
			User:        user,
			FetchStatus: entity.FetchStatusWaiting,
			OrderNumber: number,
		}

		if err := db.Create(order).Error; err != nil {
			slog.Error("failed to save order", "err", err)
			_ = c.AbortWithError(http.StatusInternalServerError, errors.New("failed to save order"))
			return
		}

		c.AbortWithStatus(http.StatusAccepted)
	}
}
