package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/models"
	mocks "github.com/NodeFactoryIo/vedran/mocks/models"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestApiController_RegisterHandler(t *testing.T) {
	// define test cases
	tests := []struct {
		name             string
		registerRequest  RegisterRequest
		httpStatus       int
		registerResponse RegisterResponse
	}{
		{
			name: "Valid registration test",
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
			// create mock controller
			nodeRepoMock := mocks.NodeRepository{}
			pingRepoMock := mocks.PingRepository{}
			nodeRepoMock.On("Save", &models.Node{
				ID:            test.registerRequest.Id,
				ConfigHash:    test.registerRequest.ConfigHash,
				NodeUrl:       test.registerRequest.NodeUrl,
				PayoutAddress: test.registerRequest.PayoutAddress,
				Token:         test.registerResponse.Token,
			}).Return(nil)
			apiController := NewApiController(&nodeRepoMock, &pingRepoMock)
			handler := http.HandlerFunc(apiController.RegisterHandler)

			// create test request
			rb, _ := json.Marshal(test.registerRequest)
			req, err := http.NewRequest("POST", "/api/v1/node", bytes.NewReader(rb))
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()

			// invoke test request
			handler.ServeHTTP(rr, req)
			var response RegisterResponse
			_ = json.Unmarshal(rr.Body.Bytes(), &response)

			// asserts
			assert.Equal(t, rr.Code, test.httpStatus, fmt.Sprintf("Response status code should be %d", test.httpStatus))
			assert.Equal(t, response, test.registerResponse, fmt.Sprintf("Response should be %v", test.registerResponse))
			assert.True(t, nodeRepoMock.AssertNumberOfCalls(t, "Save", 1))
		})
	}
	_ = os.Setenv("AUTH_SECRET", "")
}
