package controllers

import (
	"bytes"
	"context"
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	mocks "github.com/NodeFactoryIo/vedran/mocks/repositories"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestApiController_PingHandler(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "Valid ping request"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			timestamp := time.Now()
			// create mock controller
			nodeRepoMock := mocks.NodeRepository{}
			pingRepoMock := mocks.PingRepository{}
			metricsRepoMock := mocks.MetricsRepository{}
			recordRepoMock := mocks.RecordRepository{}
			pingRepoMock.On("Save", &models.Ping{
				NodeId:    "1",
				Timestamp: timestamp,
			}).Return(nil)
			apiController := NewApiController(false, repositories.Repos{
				NodeRepo:    &nodeRepoMock,
				PingRepo:    &pingRepoMock,
				MetricsRepo: &metricsRepoMock,
				RecordRepo:  &recordRepoMock,
			}, nil)
			handler := http.HandlerFunc(apiController.PingHandler)

			// create test request and populate context
			req, _ := http.NewRequest("POST", "/api/v1/node", bytes.NewReader(nil))
			c := &auth.RequestContext{
				NodeId:    "1",
				Timestamp: timestamp,
			}
			ctx := context.WithValue(req.Context(), auth.RequestContextKey, c)
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			// invoke test request
			handler.ServeHTTP(rr, req)

			// asserts
			assert.Equal(t, rr.Code, http.StatusOK, fmt.Sprintf("Response status code should be %d", http.StatusOK))
			assert.True(t, pingRepoMock.AssertNumberOfCalls(t, "Save", 1))
		})
	}
}
