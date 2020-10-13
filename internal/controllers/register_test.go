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

	"github.com/NodeFactoryIo/vedran/internal/configuration"

	"github.com/NodeFactoryIo/vedran/internal/models"
	tunnelMocks "github.com/NodeFactoryIo/vedran/mocks/http-tunnel/server"
	mocks "github.com/NodeFactoryIo/vedran/mocks/models"
	"github.com/stretchr/testify/assert"
)

func TestApiController_RegisterHandler(t *testing.T) {
	const TestTunnelURL = "test-tunnel-url:5533"
	poolMock := &tunnelMocks.Pooler{}
	configuration.Config = configuration.Configuration{
		TunnelURL: TestTunnelURL,
		PortPool:  poolMock,
	}

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
		AssignedPort                  int
		PortAssignError               error
	}{
		{
			name: "Valid registration test no whitelist",
			registerRequest: RegisterRequest{
				Id:            "1",
				ConfigHash:    "dadf2e32dwq12",
				PayoutAddress: "0xdafe2cdscdsa",
			},
			httpStatus: http.StatusOK,
			registerResponse: RegisterResponse{
				Token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdXRob3JpemVkIjp0cnVlLCJub2RlX2lkIjoiMSJ9.LdQLi-cx5HZs6HvVzSFVx0WjXFTsGqDuO9FepXfYLlY",
				TunnelURL: TestTunnelURL,
				Port:      33333,
			},
			isWhitelisted:                 false,
			saveMockReturns:               nil,
			saveMockCalledNumber:          1,
			isNodeWhitelistedMockReturns:  nil,
			isNodeWhitelistedCalledNumber: 0,
			AssignedPort:                  33333,
			PortAssignError:               nil,
		},
		{
			name: "Valid registration test nodeId on whitelist",
			registerRequest: RegisterRequest{
				Id:            "1",
				ConfigHash:    "dadf2e32dwq12",
				PayoutAddress: "0xdafe2cdscdsa",
			},
			httpStatus: http.StatusOK,
			registerResponse: RegisterResponse{
				Token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdXRob3JpemVkIjp0cnVlLCJub2RlX2lkIjoiMSJ9.LdQLi-cx5HZs6HvVzSFVx0WjXFTsGqDuO9FepXfYLlY",
				TunnelURL: TestTunnelURL,
				Port:      33333,
			},
			isWhitelisted:                 true,
			saveMockReturns:               nil,
			saveMockCalledNumber:          1,
			isNodeWhitelistedMockReturns:  nil,
			isNodeWhitelistedCalledNumber: 1,
			AssignedPort:                  33333,
			PortAssignError:               nil,
		},
		{
			name: "Invalid registration test nodeId not on whitelist",
			registerRequest: RegisterRequest{
				Id:            "1",
				ConfigHash:    "dadf2e32dwq12",
				PayoutAddress: "0xdafe2cdscdsa",
			},
			httpStatus:                    http.StatusBadRequest,
			registerResponse:              RegisterResponse{},
			isWhitelisted:                 true,
			saveMockReturns:               nil,
			saveMockCalledNumber:          0,
			isNodeWhitelistedMockReturns:  errors.New("not found"),
			isNodeWhitelistedCalledNumber: 1,
			AssignedPort:                  33333,
			PortAssignError:               nil,
		},
		{
			name: "Port assign error returns 500",
			registerRequest: RegisterRequest{
				Id:            "1",
				ConfigHash:    "dadf2e32dwq12",
				PayoutAddress: "0xdafe2cdscdsa",
			},
			httpStatus:                    http.StatusInternalServerError,
			registerResponse:              RegisterResponse{},
			isWhitelisted:                 false,
			saveMockReturns:               nil,
			saveMockCalledNumber:          0,
			isNodeWhitelistedMockReturns:  nil,
			isNodeWhitelistedCalledNumber: 0,
			AssignedPort:                  33333,
			PortAssignError:               fmt.Errorf("ERROR"),
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
				PayoutAddress: test.registerRequest.PayoutAddress,
				Token:         test.registerResponse.Token,
				LastUsed:      time.Now().Unix(),
				NodeUrl:       "http://127.0.0.1:33333",
			}).Return(test.saveMockReturns)
			poolMock.On("Acquire", "1", "1").Once().Return(test.AssignedPort, test.PortAssignError)
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
