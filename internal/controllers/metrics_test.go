package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/internal/models"
	mocks "github.com/NodeFactoryIo/vedran/mocks/models"
	"github.com/stretchr/testify/assert"
)

func TestApiController_SaveMetricsHandler(t *testing.T) {
	tests := []struct {
		name                  string
		metricsRequest        interface{}
		httpStatus            int
		repoReturn            interface{}
		numberOfRepoSaveCalls int
	}{
		{
			name: "Valid metrics save request",
			metricsRequest: MetricsRequest{
				PeerCount:             10,
				BestBlockHeight:       100,
				FinalizedBlockHeight:  100,
				ReadyTransactionCount: 10,
			},
			httpStatus:            http.StatusOK,
			repoReturn:            nil,
			numberOfRepoSaveCalls: 1,
		},
		{
			name:                  "Invalid metrics save request",
			metricsRequest:        struct{ PeerCount string }{PeerCount: "10"},
			httpStatus:            http.StatusBadRequest,
			repoReturn:            nil,
			numberOfRepoSaveCalls: 0,
		},
		{
			name: "Database fails on save",
			metricsRequest: MetricsRequest{
				PeerCount:             10,
				BestBlockHeight:       100,
				FinalizedBlockHeight:  100,
				ReadyTransactionCount: 10,
			},
			httpStatus:            http.StatusInternalServerError,
			repoReturn:            errors.New("database error"),
			numberOfRepoSaveCalls: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			timestamp := time.Now()

			// create mock controller
			nodeRepoMock := mocks.NodeRepository{}
			pingRepoMock := mocks.PingRepository{}
			metricsRepoMock := mocks.MetricsRepository{}
			recordRepoMock := mocks.RecordRepository{}
			metricsRepoMock.On("Save", &models.Metrics{
				NodeId:                "1",
				PeerCount:             10,
				BestBlockHeight:       100,
				FinalizedBlockHeight:  100,
				ReadyTransactionCount: 10,
			}).Return(test.repoReturn)
			apiController := NewApiController(false, &nodeRepoMock, &pingRepoMock, &metricsRepoMock, &recordRepoMock)
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
			assert.Equal(t, test.httpStatus, rr.Code, fmt.Sprintf("Response status code should be %d", test.httpStatus))
			assert.True(t, metricsRepoMock.AssertNumberOfCalls(t, "Save", test.numberOfRepoSaveCalls))
		})
	}
}
