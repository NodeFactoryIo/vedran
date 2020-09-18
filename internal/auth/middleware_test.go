package auth

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddleware_AuthorizedRequest(t *testing.T) {
	token, _ := CreateNewToken("test-node-id")
	req, _ := http.NewRequest("POST", "/", bytes.NewReader(nil))
	req.Header.Add("X-Auth-Token", token)
	rr := httptest.NewRecorder()

	mockHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		r := request.Context().Value("request").(*RequestContext)
		assert.Equal(t, r.NodeId, "test-node-id")
		assert.NotNil(t, r.Timestamp)
	})

	handler := AuthMiddleware(mockHandler)
	handler.ServeHTTP(rr, req)
}

func TestAuthMiddleware_UnauthorizedRequest(t *testing.T) {
	token := "invalidtokenstring"
	req, _ := http.NewRequest("POST", "/", bytes.NewReader(nil))
	req.Header.Add("X-Auth-Token", token)
	rr := httptest.NewRecorder()

	mockHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {})

	handler := AuthMiddleware(mockHandler)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusUnauthorized)
}