package handlers

import (
	"github.com/fasdalf/train-go-musthave-diploma/internal/db/entity"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"time"
)

// NewGetWithdrawalOrders gin handler
func NewGetWithdrawalOrders(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromContext(c)
		if err != nil {
			logAndAbort(c, http.StatusInternalServerError, "user in context is not a user type", err)
			return
		}
		var withdrawals []entity.Withdrawal
		query := db.Model(&entity.Withdrawal{}).Where(entity.Order{UserID: user.ID}).Order("created_at DESC")
		if err := query.Find(&withdrawals).Error; err != nil {
			logAndAbort(c, http.StatusInternalServerError, "database error", err)
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
