package handlers

import (
	"errors"
	"github.com/fasdalf/train-go-musthave-diploma/internal/db/entity"
	"github.com/fasdalf/train-go-musthave-diploma/pkg/jwt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log/slog"
	"net/http"
	"strings"
)

const currentUserKey = "currentUser"

// NewLoadUserByTokenMiddleware gin middleware
func NewLoadUserByTokenMiddleware(db *gorm.DB, key *string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Request.Header.Get(jwt.AuthHeader)
		token := strings.Replace(strings.TrimSpace(header), jwt.AuthPrefix, "", 1)
		if token == "" {
			slog.Error("missing auth header")
			_ = c.AbortWithError(http.StatusUnauthorized, errors.New("missing auth header"))
			return
		}

		userID, err := jwt.GetUserID(token, *key)
		if err != nil {
			slog.Error("invalid token", "error", err)
			_ = c.AbortWithError(http.StatusUnauthorized, errors.New("invalid token"))
			return
		}

		user := &entity.User{ID: userID}
		if err := db.Where(user).First(user).Error; err != nil {
			slog.Error("failed to find user", "err", db.Error)
			_ = c.AbortWithError(http.StatusInternalServerError, errors.New("failed to find user"))
			return
		}
		if user.ID == 0 {
			slog.Error("user does not exist")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user does not exist"))
			return
		}

		c.Set(currentUserKey, user)
		c.Next()
	}
}
