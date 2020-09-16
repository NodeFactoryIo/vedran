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

func CreateNewToken(nodeId string) (string, error) {
	// set claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"authorized": true,
		"node_id":    nodeId,
	})
	return token.SignedString([]byte(authSecret))
}
