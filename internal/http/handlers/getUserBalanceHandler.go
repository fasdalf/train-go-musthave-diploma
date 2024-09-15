package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

// GetUserBalanceHandler gin handler
func GetUserBalanceHandler(c *gin.Context) {
	user, err := getUserFromContext(c)
	if err != nil {
		slog.Error("user in context is not a user type", "error", err)
		_ = c.AbortWithError(http.StatusInternalServerError, errors.New("user in context is not a user type"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"current":   user.Amount,
		"withdrawn": user.WithdrawalsAmount,
	})
}
