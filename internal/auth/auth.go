package auth

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"os"
)

var authSecret string

func SetAuthSecret(secret string) error {
	// set auth secret if provided as arg
	if secret != "" && authSecret == "" {
		authSecret = secret
	}
	// fallback to env variable
	if authSecret == "" {
		authSecret = os.Getenv("AUTH_SECRET")
	}
	// throw error if secret not set
	if authSecret == "" {
		return errors.New("auth secret not provided")
	} else {
		return nil
	}
}

type CustomClaims struct {
	Authorized bool `json:"authorized"`
	NodeId string `json:"node_id"`
	jwt.StandardClaims
}

func CreateNewToken(nodeId string) (string, error) {
	claims := CustomClaims{
		Authorized: true,
		NodeId:     nodeId,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(authSecret))
}
