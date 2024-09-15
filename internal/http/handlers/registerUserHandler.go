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

// NewRegisterUserHandler gin handler
func NewRegisterUserHandler(db *gorm.DB, key *string) gin.HandlerFunc {
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
		if err := db.Where(user).First(user).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("failed to find user", "err", err)
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("failed to find user"))
			return
		}
		if user.ID != 0 {
			slog.Error("user already exists")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user already exists"))
			return
		}

		user.PassHash = cryptofacade.Hash(l.Password, *key)
		if err := db.Create(user).Error; err != nil {
			slog.Error("failed to create user", "err", err)
			_ = c.AbortWithError(http.StatusInternalServerError, errors.New("failed to create user"))
			return
		}

		if err := respondWithToken(c, user.ID, key); err != nil {
			slog.Error("failed to respond with jwt", "err", err)
			_ = c.AbortWithError(http.StatusInternalServerError, errors.New("failed to respond with jwt"))
			return
		}
	}
}

func respondWithToken(c *gin.Context, userID uint, key *string) error {
	token, err := jwt.BuildJWTString(userID, *key)
	if err != nil {
		return err
	}

	c.Writer.Header().Set(jwt.AuthHeader, jwt.AuthPrefix+token)
	c.JSON(http.StatusOK, gin.H{"token": token})

	return nil
}
