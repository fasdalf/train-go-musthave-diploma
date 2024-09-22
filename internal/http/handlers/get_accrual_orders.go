package handlers

import (
	"github.com/fasdalf/train-go-musthave-diploma/internal/db/entity"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"time"
)

var orderStatusMap = map[string]string{
	entity.FetchStatusWaiting:    "NEW",
	entity.FetchStatusInProgress: "PROCESSING",
	entity.FetchStatusFinished:   "PROCESSED",
	entity.FetchStatusFailure:    "INVALID",
}

// NewGetAccrualOrders gin handler
func NewGetAccrualOrders(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromContext(c)
		if err != nil {
			logAndAbort(c, http.StatusInternalServerError, "user in context is not a user type", err)
			return
		}
		var orders []entity.Order
		query := db.Model(&entity.Order{}).Where(entity.Order{UserID: user.ID}).Order("updated_at DESC")
		if err := query.Find(&orders).Error; err != nil {
			logAndAbort(c, http.StatusInternalServerError, "database error", err)
			return
		}

		result := make([]map[string]interface{}, len(orders))
		for i, order := range orders {
			// Make a map manually is faster than build new type with annotations and custom time marshaller
			row := map[string]interface{}{
				"number":      order.OrderNumber,
				"status":      orderStatusMap[order.FetchStatus],
				"uploaded_at": order.UpdatedAt.Format(time.RFC3339),
			}
			if order.Amount > 0 {
				row["accrual"] = order.Amount
			}
			result[i] = row
		}

		c.JSON(http.StatusOK, result)
	}
}
