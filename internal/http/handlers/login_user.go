package handlers

import (
	"github.com/fasdalf/train-go-musthave-diploma/internal/db/entity"
	"github.com/fasdalf/train-go-musthave-diploma/pkg/cryptofacade"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"time"
)

// NewLoginUser gin handler
func NewLoginUser(db *gorm.DB, key *string, exp time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		r := &loginPasswordRequest{}
		err := c.BindJSON(r)
		if err != nil {
			logAndAbort(c, http.StatusBadRequest, "failed to bind json", err)
			return
		}
		if err = r.validate(); err != nil {
			logAndAbort(c, http.StatusBadRequest, "invalid login or password", err)
			return
		}

		user := &entity.User{Login: r.Login}

		if err := db.Where(user).First(user).Error; err != nil {
			logAndAbort(c, http.StatusBadRequest, "failed to find user", err)
			return
		}
		if user.ID == 0 {
			logAndAbort(c, http.StatusBadRequest, "user does not exist", nil)
			return
		}

		requestPassHash := cryptofacade.Hash(r.Password, *key)
		if user.PassHash != string(requestPassHash) {
			logAndAbort(c, http.StatusBadRequest, "wrong password", err)
			return
		}

		if err := respondWithToken(c, user.ID, key, exp); err != nil {
			logAndAbort(c, http.StatusInternalServerError, "failed to respond with jwt", err)
			return
		}
	}
}
