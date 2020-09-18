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

func TestApiController_PingHandler(t *testing.T) {
	_ = os.Setenv("AUTH_SECRET", "test-auth-secret")
	timestamp := time.Now()

	// create mock controller
	nodeRepoMock := mocks.NodeRepository{}
	pingRepoMock := mocks.PingRepository{}
	pingRepoMock.On("Save", &models.Ping{
		NodeId:    "1",
		Timestamp: timestamp,
	}).Return(nil)
	apiController := NewApiController(&nodeRepoMock, &pingRepoMock)
	handler := http.HandlerFunc(apiController.PingHandler)

	// create test request and populate context
	req, _ := http.NewRequest("POST", "/api/v1/node", bytes.NewReader(nil))
	ctx := req.Context()
	ctx = context.WithValue(ctx, "node-id", "1")
	ctx = context.WithValue(ctx, "timestamp", timestamp)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	// invoke test request
	handler.ServeHTTP(rr, req)

	// asserts
	assert.Equal(t, rr.Code, http.StatusOK, fmt.Sprintf("Response status code should be %d", http.StatusOK))
	assert.True(t, pingRepoMock.AssertNumberOfCalls(t, "Save", 1))

	_ = os.Setenv("AUTH_SECRET", "")
}
