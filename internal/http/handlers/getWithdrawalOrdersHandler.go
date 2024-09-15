package handlers

import (
	"errors"
	"github.com/fasdalf/train-go-musthave-diploma/internal/db/entity"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log/slog"
	"net/http"
	"time"
)

// NewGetWithdrawalOrdersHandler gin handler
func NewGetWithdrawalOrdersHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromContext(c)
		if err != nil {
			slog.Error("user in context is not a user type", "error", err)
			_ = c.AbortWithError(http.StatusInternalServerError, errors.New("user in context is not a user type"))
			return
		}
		var withdrawals []entity.Withdrawal
		query := db.Model(&entity.Withdrawal{}).Where(entity.Order{UserID: user.ID}).Order("created_at DESC")
		if err := query.Find(&withdrawals).Error; err != nil {
			slog.Error("database error", "error", err)
			_ = c.AbortWithError(http.StatusInternalServerError, errors.New("database error"))
			return
		}
		result := make([]map[string]interface{}, len(withdrawals))
		for i, withdrawal := range withdrawals {
			// Make a map manually is faster than build new type with annotations and custom time marshaller
			row := map[string]interface{}{
				"order":        withdrawal.OrderNumber,
				"sum":          withdrawal.Amount,
				"processed_at": withdrawal.CreatedAt.Format(time.RFC3339),
			}
			result[i] = row
		}

		c.JSON(http.StatusOK, result)
	}
}
