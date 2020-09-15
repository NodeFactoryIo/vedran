package auth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/gommon/random"
	"os"
)

func CreateNewToken() (string, error) {
	// set claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"authorized": true,
	})
	return token.SignedString([]byte(getSecret()))
}

func getSecret() string {
	secret := os.Getenv("AUTH_SECRET")
	if secret == "" {
		// generate secret if not set
		generatedSecret := random.String(24, random.Alphabetic)
		_ = os.Setenv("AUTH_SECRET", generatedSecret)
		secret = generatedSecret
	}
	return secret
}