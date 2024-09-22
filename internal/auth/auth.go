package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/KirillKhitev/goph_keeper/internal/config"
	"github.com/KirillKhitev/goph_keeper/internal/models"
	"github.com/beevik/guid"
	"github.com/golang-jwt/jwt/v4"
	"strings"
	"time"
)

// TokenExp время жизни авторизационного токена.
const TokenExp = time.Hour * 3

// SecretKey секретный ключ сервера.
const SecretKey = "dswereGsdfgert2345Dsd"

// AuthorizingData хранит данные авторизации.
type AuthorizingData struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

// GenerateHashPassword генерирует хеш из пароля.
func (d *AuthorizingData) GenerateHashPassword() string {
	hashSum := GetHash(d.Password, config.ConfigServer.MasterKey)
	return hashSum
}

// NewUserFromData конструктор структуры пользователя из авторизационных данных.
func (d *AuthorizingData) NewUserFromData() models.User {
	user := models.User{
		ID:               guid.NewString(),
		UserName:         d.UserName,
		HashPassword:     d.GenerateHashPassword(),
		RegistrationDate: time.Now(),
	}

	return user
}

// Claims - требования к авторизационному токену.
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

// BuildJWTString генерирует авторизационный токен.
func BuildJWTString(user models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: user.ID,
	})

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetUserIDFromAuthHeader вытаскивает userID из авторизационного токена.
func GetUserIDFromAuthHeader(header string) (string, error) {
	tokenString := strings.TrimPrefix(header, "Bearer ")

	if tokenString == "" {
		return "", errors.New("empty authorization header")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(SecretKey), nil
		})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", fmt.Errorf("token is not valid")
	}

	return claims.UserID, nil
}

// GetHash готовит хеш из строки с помощью ключа.
func GetHash(data, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	result := h.Sum(nil)

	return hex.EncodeToString(result)
}
