package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/internal/models"
	mocks "github.com/NodeFactoryIo/vedran/mocks/models"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestApiController_SaveMetricsHandler(t *testing.T) {
	tests := []struct {
		name string
		metricsRequest MetricsRequest
		httpStatus int
	}{
		{
			name: "Valid metrics save request",
			metricsRequest: MetricsRequest{
				PeerCount:             10,
				BestBlockHeight:       100,
				FinalizedBlockHeight:  100,
				ReadyTransactionCount: 10,
			},
			httpStatus: http.StatusOK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			timestamp := time.Now()

			// create mock controller
			nodeRepoMock := mocks.NodeRepository{}
			pingRepoMock := mocks.PingRepository{}
			metricsRepoMock := mocks.MetricsRepository{}
			metricsRepoMock.On("Save", &models.Metrics{
				NodeId:                "1",
				PeerCount:             10,
				BestBlockHeight:       100,
				FinalizedBlockHeight:  100,
				ReadyTransactionCount: 10,
			}).Return(nil)
			apiController := NewApiController(&nodeRepoMock, &pingRepoMock, &metricsRepoMock)
			handler := http.HandlerFunc(apiController.SaveMetricsHandler)

			// create test request
			rb, _ := json.Marshal(test.metricsRequest)
			req, _ := http.NewRequest("POST", "/api/v1/metrics", bytes.NewReader(rb))
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
			assert.Equal(t, rr.Code, test.httpStatus, fmt.Sprintf("Response status code should be %d", test.httpStatus))
			assert.True(t, metricsRepoMock.AssertNumberOfCalls(t, "Save", 1))
		})
	}
}
