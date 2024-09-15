package handlers

import (
	"errors"
	"fmt"
	"github.com/fasdalf/train-go-musthave-diploma/internal/db/entity"
	"github.com/fasdalf/train-go-musthave-diploma/pkg/cryptofacade"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log/slog"
	"net/http"
)

// NewLoginUserHandler gin handler
func NewLoginUserHandler(db *gorm.DB, key *string) gin.HandlerFunc {
	return func(c *gin.Context) {
		l := &loginPasswordRequest{}
		err := c.BindJSON(l)
		if err != nil {
			slog.Error("failed to bind json", "err", err)
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("failed to bind json"))
			return
		}
		if err = l.validate(); err != nil {
			slog.Error("invalid login or password", "err", err)
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("failed to bind json"))
			return
		}

		user := &entity.User{Login: l.Login}

		if err := db.Where(user).First(user).Error; err != nil {
			slog.Error("failed to find user", "err", err)
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("failed to find user"))
			return
		}
		if user.ID == 0 {
			slog.Error("user does not exist")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user does not exist"))
			return
		}

		requestPassHash := cryptofacade.Hash(l.Password, *key)
		if user.PassHash != string(requestPassHash) {
			slog.Error("wrong password")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("wrong password"))
			return
		}

		if err := respondWithToken(c, user.ID, key); err != nil {
			slog.Error("failed to respond with jwt", "err", err)
			_ = c.AbortWithError(http.StatusInternalServerError, errors.New("failed to respond with jwt"))
			return
		}
	}
}

func getUserFromContext(c *gin.Context) (*entity.User, error) {
	v, ok := c.Get(currentUserKey)
	if !ok {
		return nil, fmt.Errorf("current user not found")
	}
	user, ok := v.(*entity.User)
	if !ok {
		return nil, fmt.Errorf("user in context is not a user type\"")
	}
	return user, nil
}
