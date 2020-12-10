package auth

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
)

func ParseJwtTokenWithCustomClaims(jwtToken string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(jwtToken, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(authSecret), nil
	})
}
