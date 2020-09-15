package auth

import "github.com/dgrijalva/jwt-go"

func CreateNewToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"authorized": true,
		"user_id": "1",
	})
	// TODO -> move secret to env variable
	return token.SignedString([]byte("jdnfksdmfksd"))
}