package controllers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	mocks "github.com/NodeFactoryIo/vedran/mocks/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestApiController_PingHandler(t *testing.T) {
	tests := []struct {
		name                  string
		statusCode            int
		pingSaveCallCount     int
		downtimeSaveCallCount int
		pingSaveErr           error
		downtimeSaveErr       error
		calculateDowntimeErr  error
		downtimeDuration      time.Duration
		requestTimestamp time.Time
		lastPingTimestamp time.Time
	}{
		{
			name:                  "Returns 200 if downtime calculation fails",
			statusCode:            200,
			pingSaveCallCount:     1,
			pingSaveErr:           nil,
			downtimeSaveErr:       nil,
			downtimeSaveCallCount: 0,
			downtimeDuration:      time.Duration(0),
			calculateDowntimeErr:  fmt.Errorf("ERROR"),
			requestTimestamp: time.Now(),
			lastPingTimestamp: time.Now().Add(-5 * time.Second),
		},
		{
			name:                  "Returns 200 if downtime save fails",
			statusCode:            200,
			pingSaveCallCount:     1,
			pingSaveErr:           nil,
			downtimeSaveErr:       fmt.Errorf("ERROR"),
			downtimeSaveCallCount: 1,
			downtimeDuration:      time.Duration(time.Second * 19),
			calculateDowntimeErr:  nil,
			requestTimestamp: time.Now(),
			lastPingTimestamp: time.Now().Add(-19 * time.Second),
		},
		{
			name:                  "Saves downtime if downtime duration more than 18 seconds",
			statusCode:            200,
			pingSaveCallCount:     1,
			pingSaveErr:           nil,
			downtimeSaveErr:       nil,
			downtimeSaveCallCount: 1,
			downtimeDuration:      time.Duration(time.Second * 19),
			calculateDowntimeErr:  nil,
			requestTimestamp: time.Now(),
			lastPingTimestamp: time.Now().Add(-19 * time.Second),
		},
		{
			name:                  "Returns 500 if saving ping fails",
			statusCode:            500,
			pingSaveCallCount:     1,
			pingSaveErr:           fmt.Errorf("ERROR"),
			downtimeSaveErr:       nil,
			downtimeSaveCallCount: 0,
			downtimeDuration:      time.Duration(time.Second * 8),
			calculateDowntimeErr:  nil,
			requestTimestamp: time.Now(),
			lastPingTimestamp: time.Now().Add(-5 * time.Second),
		},
		{
			name:                  "Returns 200 and does not save downtime if downtime duration less than 5 + 5 seconds",
			statusCode:            200,
			pingSaveCallCount:     1,
			pingSaveErr:           nil,
			downtimeSaveErr:       nil,
			downtimeSaveCallCount: 0,
			downtimeDuration:      time.Duration(time.Second * 8),
			calculateDowntimeErr:  nil,
			requestTimestamp: time.Now(),
			lastPingTimestamp: time.Now().Add(-5 * time.Second),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// create mock controller
			nodeRepoMock := mocks.NodeRepository{}
			recordRepoMock := mocks.RecordRepository{}
			metricsRepoMock := mocks.MetricsRepository{}

			pingRepoMock := mocks.PingRepository{}
			pingRepoMock.On("Save", &models.Ping{
				NodeId:    "1",
				Timestamp: test.requestTimestamp,
			}).Return(test.pingSaveErr)
			pingRepoMock.On("CalculateDowntime", mock.Anything, mock.Anything).Return(
				test.lastPingTimestamp, test.downtimeDuration, test.calculateDowntimeErr)

			downtimeRepoMock := mocks.DowntimeRepository{}
			downtimeRepoMock.On("Save", mock.Anything).Return(test.downtimeSaveErr)

			apiController := NewApiController(false, repositories.Repos{
				NodeRepo:     &nodeRepoMock,
				PingRepo:     &pingRepoMock,
				MetricsRepo:  &metricsRepoMock,
				RecordRepo:   &recordRepoMock,
				DowntimeRepo: &downtimeRepoMock,
			}, nil)
			handler := http.HandlerFunc(apiController.PingHandler)

			// create test request and populate context
			req, _ := http.NewRequest("POST", "/api/v1/node", bytes.NewReader(nil))
			c := &auth.RequestContext{
				NodeId:    "1",
				Timestamp: test.requestTimestamp,
			}
			ctx := context.WithValue(req.Context(), auth.RequestContextKey, c)
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			// invoke test request
			handler.ServeHTTP(rr, req)

			// asserts
			assert.Equal(t, rr.Code, test.statusCode, fmt.Sprintf("Response status code should be %d", test.statusCode))
			assert.True(t, pingRepoMock.AssertNumberOfCalls(t, "Save", test.pingSaveCallCount))
			assert.True(t, downtimeRepoMock.AssertNumberOfCalls(t, "Save", test.downtimeSaveCallCount))
		})
	}
}
