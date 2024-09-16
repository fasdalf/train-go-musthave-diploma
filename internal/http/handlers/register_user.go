package handlers

import (
	"errors"
	"github.com/fasdalf/train-go-musthave-diploma/internal/db/entity"
	"github.com/fasdalf/train-go-musthave-diploma/pkg/cryptofacade"
	"github.com/fasdalf/train-go-musthave-diploma/pkg/jwt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log/slog"
	"net/http"
	"time"
)

type loginPasswordRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (r *loginPasswordRequest) validate() error {
	if r.Login == "" || r.Password == "" {
		return errors.New("login and password must be not empty")
	}
	return nil
}

// NewRegisterUser gin handler
func NewRegisterUser(db *gorm.DB, key *string, exp time.Duration) gin.HandlerFunc {
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
		if err := db.Where(user).First(user).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logAndAbort(c, http.StatusBadRequest, "failed to find user", err)
			return
		}
		if user.ID != 0 {
			logAndAbort(c, http.StatusBadRequest, "user already exists", nil)
			return
		}

		user.PassHash = cryptofacade.Hash(r.Password, *key)
		if err := db.Create(user).Error; err != nil {
			logAndAbort(c, http.StatusInternalServerError, "failed to create user", err)
			return
		}

		if err := respondWithToken(c, user.ID, key, exp); err != nil {
			logAndAbort(c, http.StatusInternalServerError, "failed to respond with jwt", err)
			return
		}
	}
}

func respondWithToken(c *gin.Context, userID uint, key *string, exp time.Duration) error {
	token, err := jwt.BuildJWTString(userID, *key, exp)
	if err != nil {
		return err
	}

	c.Writer.Header().Set(jwt.AuthHeader, jwt.AuthPrefix+token)
	c.JSON(http.StatusOK, gin.H{"token": token})

	return nil
}

// logAndAbort - abort context. Use literal as message to be able to find a source of the error.
func logAndAbort(c *gin.Context, code int, message string, err error) {
	slog.Error(message, "err", err)
	_ = c.AbortWithError(code, errors.New(message))
}
