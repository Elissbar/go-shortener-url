package common

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func GenerateAuthToken(jwtSecret string) (*http.Cookie, string, error) {
	userID, err := uuid.NewRandom()
	if err != nil {
		return &http.Cookie{}, "", err
	}
	userIDStr := userID.String()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{UserID: userIDStr})
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return &http.Cookie{}, "", err
	}

	cookie := &http.Cookie{
		Name:     "user_id",
		Value:    tokenString,
		HttpOnly: true,
	}
	return cookie, userIDStr, nil
}

func VerifyAuthToken(tokenString, jwtSecret string) (string, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return "", fmt.Errorf("token parsing failed: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	if claims.UserID == "" {
		return "", fmt.Errorf("user_id is empty")
	}

	return claims.UserID, nil
}
