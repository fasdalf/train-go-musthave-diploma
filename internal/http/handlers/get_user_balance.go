package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetUserBalance gin handler
func GetUserBalance(c *gin.Context) {
	user, err := getUserFromContext(c)
	if err != nil {
		logAndAbort(c, http.StatusInternalServerError, "user in context is not a user type", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"current":   user.Amount,
		"withdrawn": user.WithdrawalsAmount,
	})
}
