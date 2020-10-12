package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/NodeFactoryIo/vedran/internal/models"
	mocks "github.com/NodeFactoryIo/vedran/mocks/models"
	"github.com/stretchr/testify/assert"
)

func TestApiController_RegisterHandler(t *testing.T) {
	// define test cases
	tests := []struct {
		name                          string
		registerRequest               RegisterRequest
		httpStatus                    int
		registerResponse              RegisterResponse
		isWhitelisted                 bool
		saveMockReturns               interface{}
		saveMockCalledNumber          int
		isNodeWhitelistedMockReturns  interface{}
		isNodeWhitelistedCalledNumber int
	}{
		{
			name: "Valid registration test no whitelist",
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
			isWhitelisted:                 false,
			saveMockReturns:               nil,
			saveMockCalledNumber:          1,
			isNodeWhitelistedMockReturns:  nil,
			isNodeWhitelistedCalledNumber: 0,
		},
		{
			name: "Valid registration test nodeId on whitelist",
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
			isWhitelisted:                 true,
			saveMockReturns:               nil,
			saveMockCalledNumber:          1,
			isNodeWhitelistedMockReturns:  nil,
			isNodeWhitelistedCalledNumber: 1,
		},
		{
			name: "Invalid registration test nodeId not on whitelist",
			registerRequest: RegisterRequest{
				Id:            "1",
				ConfigHash:    "dadf2e32dwq12",
				NodeUrl:       "node.test.url",
				PayoutAddress: "0xdafe2cdscdsa",
			},
			httpStatus:                    http.StatusBadRequest,
			registerResponse:              RegisterResponse{},
			isWhitelisted:                 true,
			saveMockReturns:               nil,
			saveMockCalledNumber:          0,
			isNodeWhitelistedMockReturns:  errors.New("not found"),
			isNodeWhitelistedCalledNumber: 1,
		},
	}
	_ = os.Setenv("AUTH_SECRET", "test-auth-secret")
	// execute tests
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create mock controller
			nodeRepoMock := mocks.NodeRepository{}
			pingRepoMock := mocks.PingRepository{}
			metricsRepoMock := mocks.MetricsRepository{}
			recordRepoMock := mocks.RecordRepository{}
			nodeRepoMock.On("Save", &models.Node{
				ID:            test.registerRequest.Id,
				ConfigHash:    test.registerRequest.ConfigHash,
				NodeUrl:       test.registerRequest.NodeUrl,
				PayoutAddress: test.registerRequest.PayoutAddress,
				Token:         test.registerResponse.Token,
				LastUsed:      time.Now().Unix(),
			}).Return(test.saveMockReturns)
			nodeRepoMock.On("IsNodeWhitelisted", test.registerRequest.Id).Return(true, test.isNodeWhitelistedMockReturns)
			apiController := NewApiController(test.isWhitelisted, &nodeRepoMock, &pingRepoMock, &metricsRepoMock, &recordRepoMock)
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

			// asserts
			assert.Equal(t, rr.Code, test.httpStatus, fmt.Sprintf("Response status code should be %d", test.httpStatus))

			var response RegisterResponse
			if rr.Code == http.StatusOK {
				_ = json.Unmarshal(rr.Body.Bytes(), &response)
				assert.Equal(t, response, test.registerResponse, fmt.Sprintf("Response should be %v", test.registerResponse))
			}

			assert.True(t, nodeRepoMock.AssertNumberOfCalls(t, "Save", test.saveMockCalledNumber))
			assert.True(t, nodeRepoMock.AssertNumberOfCalls(t, "IsNodeWhitelisted", test.isNodeWhitelistedCalledNumber))
		})
	}
	_ = os.Setenv("AUTH_SECRET", "")
}
