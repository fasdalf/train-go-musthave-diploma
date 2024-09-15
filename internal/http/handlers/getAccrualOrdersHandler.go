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

var orderStatusMap = map[string]string{
	entity.FetchStatusWaiting:    "NEW",
	entity.FetchStatusInProgress: "PROCESSING",
	entity.FetchStatusFinished:   "PROCESSED",
	entity.FetchStatusFailure:    "INVALID",
}

// NewGetAccrualOrdersHandler gin handler
func NewGetAccrualOrdersHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromContext(c)
		if err != nil {
			slog.Error("user in context is not a user type", "error", err)
			_ = c.AbortWithError(http.StatusInternalServerError, errors.New("user in context is not a user type"))
			return
		}
		var orders []entity.Order
		query := db.Model(&entity.Order{}).Where(entity.Order{UserID: user.ID}).Order("updated_at DESC")
		if err := query.Find(&orders).Error; err != nil {
			slog.Error("database error", "error", err)
			_ = c.AbortWithError(http.StatusInternalServerError, errors.New("database error"))
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
