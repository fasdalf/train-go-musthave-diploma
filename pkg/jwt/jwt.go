package jwt

import (
	"fmt"
	jwt "github.com/golang-jwt/jwt/v4"
	"time"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID uint
}

const (
	TokenExp   = time.Hour * 3
	AuthHeader = "Authorization"
	AuthPrefix = "Bearer "
)

// BuildJWTString создаёт токен и возвращает его в виде строки
func BuildJWTString(id uint, key string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда истекает токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		// собственное утверждение
		UserID: id,
	})

	tokenString, err := token.SignedString([]byte(key))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetUserID валидирует токен и возвращает ID пользователя
func GetUserID(tokenString, key string) (uint, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Method.Alg())
		}
		return []byte(key), nil
	})
	if err != nil {
		return 0, err
	}

	return claims.UserID, nil
}
