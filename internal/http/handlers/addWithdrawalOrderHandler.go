package handlers

import (
	"errors"
	"github.com/fasdalf/train-go-musthave-diploma/internal/db/entity"
	"github.com/fasdalf/train-go-musthave-diploma/pkg/luhn"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log/slog"
	"net/http"
)

type withdrawRequest struct {
	Order string
	Sum   float64
}

// NewAddWithdrawalOrderHandler gin handler
func NewAddWithdrawalOrderHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromContext(c)
		if err != nil {
			slog.Error("user in context is not a user type", "error", err)
			_ = c.AbortWithError(http.StatusInternalServerError, errors.New("user in context is not a user type"))
			return
		}

		req := &withdrawRequest{}
		err = c.BindJSON(req)
		if err != nil {
			slog.Error("json parse error", "error", err)
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("json parse error"))
			return
		}

		if req.Sum <= 0 {
			slog.Error("sum is not positive")
			_ = c.AbortWithError(http.StatusUnprocessableEntity, errors.New("sum is not positive"))
			return
		}

		if !luhn.LuhnAlgorithm(req.Order) {
			slog.Error("number is mystyped")
			_ = c.AbortWithError(http.StatusUnprocessableEntity, errors.New("number is mystyped"))
			return
		}

		// start a TX
		tx := db.WithContext(c).Begin()
		// check number in orders
		var count int
		err = tx.Raw("SELECT COUNT(ID) FROM orders WHERE order_number = ?", req.Order).Scan(&count).Error
		if err != nil {
			slog.Error("database error", "error", err)
			_ = c.AbortWithError(http.StatusInternalServerError, errors.New("database error"))
			tx.Rollback()
			return
		}
		if count > 0 {
			slog.Error("number is used", "error", err)
			_ = c.AbortWithError(http.StatusUnprocessableEntity, errors.New("number is used"))
			tx.Rollback()
			return
		}
		// check number in withdrawals
		err = tx.Raw("SELECT COUNT(ID) FROM withdrawals WHERE order_number = ?", req.Order).Scan(&count).Error
		if err != nil {
			slog.Error("database error", "error", err)
			_ = c.AbortWithError(http.StatusInternalServerError, errors.New("database error"))
			tx.Rollback()
			return
		}
		if count > 0 {
			slog.Error("number is used", "error", err)
			_ = c.AbortWithError(http.StatusUnprocessableEntity, errors.New("number is used"))
			tx.Rollback()
			return
		}
		// lock user for update
		accountUser := &entity.User{ID: user.ID}
		query := tx.Model(&entity.User{}).Where(accountUser).Clauses(clause.Locking{Strength: "UPDATE"})
		if err := query.Find(&accountUser).Error; err != nil {
			slog.Error("database error", "error", err)
			_ = c.AbortWithError(http.StatusInternalServerError, errors.New("database error"))
			tx.Rollback()
			return
		}
		if accountUser.ID == 0 {
			slog.Error("can't fetch user balance")
			_ = c.AbortWithError(http.StatusInternalServerError, errors.New("can't fetch user balance"))
			tx.Rollback()
			return
		}
		if accountUser.Amount < req.Sum {
			slog.Error("balance amount is insufficent")
			_ = c.AbortWithError(http.StatusPaymentRequired, errors.New("can't fetch user balance"))
			tx.Rollback()
			return
		}
		// save withdraw
		withdraw := &entity.Withdrawal{
			UserID:      user.ID,
			OrderNumber: req.Order,
			Amount:      req.Sum,
		}
		tx.Save(withdraw)
		if err := tx.Save(withdraw).Error; err != nil {
			slog.Error("database error", "error", err)
			_ = c.AbortWithError(http.StatusInternalServerError, errors.New("database error"))
			tx.Rollback()
			return
		}
		// modify user's amount
		if err := tx.Exec(
			"UPDATE users SET amount = amount - ?, withdrawals_amount = withdrawals_amount + ? WHERE id = ?",
			req.Sum,
			req.Sum,
			user.ID,
		).Error; err != nil {
			slog.Error("database error", "error", err)
			_ = c.AbortWithError(http.StatusInternalServerError, errors.New("database error"))
			tx.Rollback()
			return
		}
		// commit
		if err := tx.Commit().Error; err != nil {
			slog.Error("database error", "error", err)
			_ = c.AbortWithError(http.StatusInternalServerError, errors.New("database error"))
			return
		}
		// respond 200
		c.AbortWithStatus(http.StatusOK)
	}
}
