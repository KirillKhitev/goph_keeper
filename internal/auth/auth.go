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

const TokenExp = time.Hour * 3
const SecretKey = "dswereGsdfgert2345Dsd"

type AuthorizingData struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

func (d *AuthorizingData) GenerateHashPassword() string {
	hashSum := GetHash(d.Password, config.ConfigServer.MasterKey)
	return hashSum
}

func (d *AuthorizingData) NewUserFromData() models.User {
	user := models.User{
		ID:               guid.NewString(),
		UserName:         d.UserName,
		HashPassword:     d.GenerateHashPassword(),
		RegistrationDate: time.Now(),
	}

	return user
}

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

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

func GetHash(data, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	result := h.Sum(nil)

	return hex.EncodeToString(result)
}
