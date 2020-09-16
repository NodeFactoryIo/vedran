package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/magiconair/properties/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRegisterHandler(t *testing.T) {
	// define test cases
	tests := []struct{
		name string
		registerRequest RegisterRequest
		httpStatus int
		registerResponse RegisterResponse
	}{
		{
			name:            "Valid registration test",
			registerRequest: RegisterRequest{
				Id:            "1",
				ConfigHash:    "dadf2e32dwq12",
				NodeUrl:       "node.test.url",
				PayoutAddress: "0xdafe2cdscdsa",
			},
			httpStatus: http.StatusOK,
			registerResponse: RegisterResponse{
				Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdXRob3JpemVkIjp0cnVlLCJub2RlX2lkIjoiMSJ9.LdQLi-cx5HZs6HvVzSFVx0WjXFTsGqDuO9FepXfYLlY",
			},
		},
	}
	_ = os.Setenv("AUTH_SECRET", "test-auth-secret")
	// execute tests
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create test request
			rb, _ := json.Marshal(test.registerRequest)
			req, err := http.NewRequest("POST", "/api/v1/node", bytes.NewReader(rb))
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(RegisterHandler)
			// invoke test request
			handler.ServeHTTP(rr, req)
			var response RegisterResponse
			_ = json.Unmarshal(rr.Body.Bytes(), &response)
			// asserts
			assert.Equal(t, rr.Code, test.httpStatus, fmt.Sprintf("Response status code should be %d", test.httpStatus))
			assert.Equal(t, response, test.registerResponse, fmt.Sprintf("Response should be %v", test.registerResponse))
		})
	}
	_ = os.Setenv("AUTH_SECRET", "")
}