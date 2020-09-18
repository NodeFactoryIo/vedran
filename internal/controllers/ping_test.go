package controllers

import (
	"bytes"
	"context"
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/models"
	mocks "github.com/NodeFactoryIo/vedran/mocks/models"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestPingHandler(t *testing.T) {
	// define test cases
	tests := []struct {
		name             string
		httpStatus       int
	}{
		{
			name: "Valid ping test",
			httpStatus: http.StatusOK,
		},
	}
	_ = os.Setenv("AUTH_SECRET", "test-auth-secret")
	// execute tests
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			timestamp := time.Now()
			nodeRepoMock := mocks.NodeRepository{}
			pingRepoMock := mocks.PingRepository{}
			pingRepoMock.On("Save", &models.Ping{
				NodeId:    "1",
				Timestamp: timestamp,
			}).Return(nil)

			rr := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/node", bytes.NewReader(nil))
			ctx := req.Context()
			ctx = context.WithValue(ctx, "node-id", "1")
			ctx = context.WithValue(ctx, "timestamp", timestamp)
			req = req.WithContext(ctx)

			apiController := NewApiController(&nodeRepoMock, &pingRepoMock)
			handler := http.HandlerFunc(apiController.PingHandler)
			// invoke test request
			handler.ServeHTTP(rr, req)

			// asserts
			assert.Equal(t, rr.Code, test.httpStatus, fmt.Sprintf("Response status code should be %d", test.httpStatus))
			assert.True(t, pingRepoMock.AssertNumberOfCalls(t, "Save", 1))
		})
	}
	_ = os.Setenv("AUTH_SECRET", "")
}