package auth

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	validToken, _ := CreateNewToken("test-node-id")
	tests := []struct {
		name string
		token string
		status int
		mockHandle http.HandlerFunc
	}{
		{
			name: "Authorized request",
			token: validToken,
			status: http.StatusOK,
			mockHandle: func(writer http.ResponseWriter, request *http.Request) {
				r := request.Context().Value(RequestContextKey).(*RequestContext)
				assert.Equal(t, r.NodeId, "test-node-id")
				assert.NotNil(t, r.Timestamp)
			},
		},
		{
			name: "Unauthorized request",
			token: "invalidtokenstring",
			status: http.StatusUnauthorized,
			mockHandle: func(writer http.ResponseWriter, request *http.Request) {},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/", bytes.NewReader(nil))
			req.Header.Add("X-Auth-Header", test.token)
			rr := httptest.NewRecorder()

			handler := AuthMiddleware(test.mockHandle)
			handler.ServeHTTP(rr, req)
			assert.Equal(t, rr.Code, test.status)
		})
	}
}
