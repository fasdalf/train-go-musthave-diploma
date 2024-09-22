package handlers

import (
	"github.com/fasdalf/train-go-musthave-diploma/internal/db/entity"
	"github.com/fasdalf/train-go-musthave-diploma/pkg/luhn"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/http"
)

type withdrawRequest struct {
	Order string
	Sum   float64
}

// NewAddWithdrawalOrder gin handler
func NewAddWithdrawalOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromContext(c)
		if err != nil {
			logAndAbort(c, http.StatusInternalServerError, "user in context is not a user type", err)
			return
		}

		req := &withdrawRequest{}
		err = c.BindJSON(req)
		if err != nil {
			logAndAbort(c, http.StatusBadRequest, "json parse error", err)
			return
		}

		if req.Sum <= 0 {
			logAndAbort(c, http.StatusUnprocessableEntity, "sum is not positive", nil)
			return
		}

		if !luhn.LuhnAlgorithm(req.Order) {
			logAndAbort(c, http.StatusUnprocessableEntity, "number is mystyped", err)
			return
		}

		// start a TX
		tx := db.WithContext(c).Begin()
		errorsCount := len(c.Errors)
		saveWithdrawal(c, tx, req, user)
		if len(c.Errors) != errorsCount {
			tx.Rollback()
			return
		}
		// commit
		if err := tx.Commit().Error; err != nil {
			logAndAbort(c, http.StatusInternalServerError, "database error", err)
			return
		}
		// respond 200
		c.AbortWithStatus(http.StatusOK)
	}
}

func saveWithdrawal(c *gin.Context, tx *gorm.DB, req *withdrawRequest, user *entity.User) {
	// check number in orders
	var count int
	err := tx.Raw("SELECT COUNT(ID) FROM orders WHERE order_number = ?", req.Order).Scan(&count).Error
	if err != nil {
		logAndAbort(c, http.StatusInternalServerError, "database error", err)
		return
	}
	if count > 0 {
		logAndAbort(c, http.StatusUnprocessableEntity, "number is used", nil)
		return
	}
	// check number in withdrawals
	err = tx.Raw("SELECT COUNT(ID) FROM withdrawals WHERE order_number = ?", req.Order).Scan(&count).Error
	if err != nil {
		logAndAbort(c, http.StatusInternalServerError, "database error", err)
		return
	}
	if count > 0 {
		logAndAbort(c, http.StatusUnprocessableEntity, "number is used", err)
		return
	}
	// lock user for update
	accountUser := &entity.User{ID: user.ID}
	query := tx.Model(&entity.User{}).Where(accountUser).Clauses(clause.Locking{Strength: "UPDATE"})
	if err := query.Find(&accountUser).Error; err != nil {
		logAndAbort(c, http.StatusInternalServerError, "database error", err)
		return
	}
	if accountUser.ID == 0 {
		logAndAbort(c, http.StatusInternalServerError, "can't fetch user balance", err)
		return
	}
	if accountUser.Amount < req.Sum {
		logAndAbort(c, http.StatusPaymentRequired, "balance amount is insufficent", err)
		return
	}
	// save withdraw
	withdraw := &entity.Withdrawal{
		UserID:      user.ID,
		OrderNumber: req.Order,
		Amount:      req.Sum,
	}
	if err := tx.Save(withdraw).Error; err != nil {
		logAndAbort(c, http.StatusInternalServerError, "database error", err)
		return
	}
	// modify user's amount
	if err := tx.Exec(
		"UPDATE users SET amount = amount - ?, withdrawals_amount = withdrawals_amount + ? WHERE id = ?",
		req.Sum,
		req.Sum,
		user.ID,
	).Error; err != nil {
		logAndAbort(c, http.StatusInternalServerError, "database error", err)
		return
	}
}
