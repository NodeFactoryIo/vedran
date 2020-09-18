package auth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSetAuthSecret(t *testing.T) {
	tests := []struct {
		name       string
		argument   string
		env        string
		shouldFail bool
	}{
		{name: "Auth secret as param", argument: "auth-secret", env: "", shouldFail: false},
		{name: "Auth secret as env variable", argument: "", env: "auth-secret", shouldFail: false},
		{name: "Auth secret not set", argument: "", env: "", shouldFail: true},
		{name: "Auth secret as param has priority", argument: "auth-secret", env: "auth-secret-env", shouldFail: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.env != "" {
				_ = os.Setenv("AUTH_SECRET", test.env)
			}
			err := SetAuthSecret(test.argument)
			// assert error
			if test.shouldFail {
				assert.Error(t, err, "Setting auth secret should generate error")
				assert.Equal(t, authSecret, "", "Auth secret should not be set")
			} else {
				assert.NoError(t, err, "Setting auth secret should not generate error")
				assert.Equal(t, authSecret, "auth-secret", "Auth secret should be set")
			}

			// reset auth secret and env variable
			authSecret = ""
			_ = os.Unsetenv("AUTH_SECRET")
		})
	}
}

func TestCreateNewToken(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "Valid token with claims generated"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jwtToken, err := CreateNewToken("test-node-1")
			assert.NoError(t, err, "Should successfully generate token")
			token, err := jwt.ParseWithClaims(jwtToken, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(authSecret), nil
			})
			assert.NoError(t, err, "Should successfully parse token")
			claims, ok := token.Claims.(*CustomClaims)
			assert.True(t, ok, "Should contain custom claims")
			assert.Equal(t, "test-node-1", claims.NodeId, "Claims should have nodeId")
		})
	}
}

