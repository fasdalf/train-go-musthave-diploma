package handlers

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/fasdalf/train-go-musthave-diploma/internal/db/entity"
	internalErrors "github.com/fasdalf/train-go-musthave-diploma/internal/errors"
	"github.com/fasdalf/train-go-musthave-diploma/pkg/luhn"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"net/http"
	"strings"
)

// NewAddAccrualOrder gin handler
func NewAddAccrualOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		number, err := getOrderNumber(c)
		if err != nil {
			code := http.StatusBadRequest
			if errors.Is(err, internalErrors.ErrNotRegistered) {
				code = http.StatusUnprocessableEntity
			}
			logAndAbort(c, code, "number is not provided", err)
			return
		}

		order := &entity.Order{OrderNumber: number}
		if err := db.Where(order).Find(order).Error; err != nil {
			logAndAbort(c, http.StatusBadRequest, "failed to find order", err)
			return
		}

		user, err := getUserFromContext(c)
		if err != nil {
			logAndAbort(c, http.StatusInternalServerError, "user in context is not a user type", err)
			return
		}

		if order.ID != 0 {
			if order.UserID == user.ID {
				c.JSON(http.StatusOK, gin.H{"status": order.FetchStatus})
				return
			}
			logAndAbort(c, http.StatusConflict, "order already exists", err)
			return
		}

		order = &entity.Order{
			User:        user,
			FetchStatus: entity.FetchStatusWaiting,
			OrderNumber: number,
		}

		if err := db.Create(order).Error; err != nil {
			logAndAbort(c, http.StatusInternalServerError, "failed to save order", err)
			return
		}

		c.AbortWithStatus(http.StatusAccepted)
	}
}

func getOrderNumber(c *gin.Context) (string, error) {
	b := bufio.NewReader(c.Request.Body)
	s, err := b.ReadString(0)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	number := strings.TrimSpace(s)
	if len(number) == 0 {
		logAndAbort(c, http.StatusBadRequest, "number is whitespace", err)
		return "", errors.New("number is whitespace")
	}
	if !luhn.LuhnAlgorithm(number) {
		logAndAbort(c, http.StatusUnprocessableEntity, "number is mystyped", nil)
		return "", fmt.Errorf("%w number is mystyped", internalErrors.ErrUnprocessableEntity)
		//return "", fmt.Errorf("%w number is mystyped", catchable.ErrNotRegistered)
	}

	return number, nil
}
