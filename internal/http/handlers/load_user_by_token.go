package handlers

import (
	"fmt"
	"github.com/fasdalf/train-go-musthave-diploma/internal/db/entity"
	"github.com/fasdalf/train-go-musthave-diploma/pkg/jwt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

const currentUserKey = "currentUser"

// NewLoadUserByToken gin auth middleware
func NewLoadUserByToken(db *gorm.DB, key *string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Request.Header.Get(jwt.AuthHeader)
		token := strings.Replace(strings.TrimSpace(header), jwt.AuthPrefix, "", 1)
		if token == "" {
			logAndAbort(c, http.StatusUnauthorized, "missing auth header", nil)
			return
		}

		userID, err := jwt.GetUserID(token, *key)
		if err != nil {
			logAndAbort(c, http.StatusUnauthorized, "invalid token", err)
			return
		}

		user := &entity.User{ID: userID}
		if err := db.Where(user).First(user).Error; err != nil {
			logAndAbort(c, http.StatusInternalServerError, "failed to find user", err)
			return
		}
		if user.ID == 0 {
			logAndAbort(c, http.StatusBadRequest, "user does not exist", nil)
			return
		}

		c.Set(currentUserKey, user)
		c.Next()
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
