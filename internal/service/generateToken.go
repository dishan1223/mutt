package service

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var JwtSecret string

func MustInitJWT(s string) {
	if s == "" {
		panic("PANIC :: JwtSecret is not set.")
	}
	JwtSecret = s
}

func GenerateToken(userId uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userId,
		"exp":     time.Now().Add(time.Hour * 24 * 365).Unix(), // Token expires in 7 days
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JwtSecret))
}
