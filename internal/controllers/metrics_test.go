package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	mocks "github.com/NodeFactoryIo/vedran/mocks/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestApiController_SaveMetricsHandler(t *testing.T) {
	tests := []struct {
		name           string
		metricsRequest interface{}
		nodeId         string
		httpStatus     int
		// NodeRepo.FindByID
		nodeRepoIsNodeOnCooldownReturns bool
		nodeRepoIsNodeOnCooldownError   error
		nodeRepoIsNodeOnNumOfCalls      int
		// NodeRepo.AddNodeToActive
		nodeRepoAddNodeToActiveError      error
		nodeRepoAddNodeToActiveNumOfCalls int
		// MetricsRepo.FindByID
		metricsRepoFindByIDError      error
		metricsRepoFindByIDReturn     *models.Metrics
		metricsRepoFindByIDNumOfCalls int
		// MetricsRepo.GetLatestBlockMetrics
		metricsRepoGetLatestBlockMetricsError      error
		metricsRepoGetLatestBlockMetricsReturn     *models.LatestBlockMetrics
		metricsRepoGetLatestBlockMetricsNumOfCalls int
		// MetricsRepo.Save
		metricsRepoSaveError      error
		metricsRepoSaveNumOfCalls int
	}{
		{
			name: "Valid metrics save request and node should be added to active nodes",
			metricsRequest: MetricsRequest{
				PeerCount:             0,
				BestBlockHeight:       1000,
				FinalizedBlockHeight:  995,
				ReadyTransactionCount: 0,
			},
			nodeId:     "1",
			httpStatus: http.StatusOK,
			// NodeRepo.FindByID
			nodeRepoIsNodeOnCooldownReturns: false,
			nodeRepoIsNodeOnCooldownError: nil,
			nodeRepoIsNodeOnNumOfCalls:    1,
			// NodeRepo.AddNodeToActive
			nodeRepoAddNodeToActiveError:      nil,
			nodeRepoAddNodeToActiveNumOfCalls: 1,
			// MetricsRepo.FindByID
			metricsRepoFindByIDReturn: &models.Metrics{
				NodeId:                "1",
				PeerCount:             0,
				BestBlockHeight:       1000,
				FinalizedBlockHeight:  995,
				ReadyTransactionCount: 0,
			},
			metricsRepoFindByIDError:      nil,
			metricsRepoFindByIDNumOfCalls: 1,
			// MetricsRepo.GetLatestBlockMetrics
			metricsRepoGetLatestBlockMetricsReturn: &models.LatestBlockMetrics{
				BestBlockHeight:      1001,
				FinalizedBlockHeight: 998,
			},
			metricsRepoGetLatestBlockMetricsError:      nil,
			metricsRepoGetLatestBlockMetricsNumOfCalls: 1,
			// MetricsRepo.Save
			metricsRepoSaveError:      nil,
			metricsRepoSaveNumOfCalls: 1,
		},
		{
			name: "Valid metrics save request and node should not be added to active nodes as it is penalized",
			metricsRequest: MetricsRequest{
				PeerCount:             10,
				BestBlockHeight:       100,
				FinalizedBlockHeight:  100,
				ReadyTransactionCount: 10,
			},
			nodeId:     "1",
			httpStatus: http.StatusOK,
			// NodeRepo.FindByID
			nodeRepoIsNodeOnCooldownReturns: true,
			nodeRepoIsNodeOnCooldownError:      nil,
			nodeRepoIsNodeOnNumOfCalls: 1,
			// NodeRepo.AddNodeToActive
			nodeRepoAddNodeToActiveError:      nil,
			nodeRepoAddNodeToActiveNumOfCalls: 0,
			// MetricsRepo.FindByID
			metricsRepoFindByIDReturn:     nil,
			metricsRepoFindByIDError:      nil,
			metricsRepoFindByIDNumOfCalls: 0,
			// MetricsRepo.GetLatestBlockMetrics
			metricsRepoGetLatestBlockMetricsReturn:     nil,
			metricsRepoGetLatestBlockMetricsError:      nil,
			metricsRepoGetLatestBlockMetricsNumOfCalls: 0,
			// MetricsRepo.Save
			metricsRepoSaveError:      nil,
			metricsRepoSaveNumOfCalls: 1,
		},
		{
			name: "Valid metrics save request and node should not be added to active nodes as metrics are invalid",
			metricsRequest: MetricsRequest{
				PeerCount:             0,
				BestBlockHeight:       1000,
				FinalizedBlockHeight:  995,
				ReadyTransactionCount: 0,
			},
			nodeId:     "1",
			httpStatus: http.StatusOK,
			// NodeRepo.FindByID
			nodeRepoIsNodeOnCooldownReturns: false,
			nodeRepoIsNodeOnCooldownError:      nil,
			nodeRepoIsNodeOnNumOfCalls: 1,
			// NodeRepo.AddNodeToActive
			nodeRepoAddNodeToActiveError:      nil,
			nodeRepoAddNodeToActiveNumOfCalls: 0,
			// MetricsRepo.FindByID
			metricsRepoFindByIDReturn: &models.Metrics{
				NodeId:                "1",
				PeerCount:             0,
				BestBlockHeight:       1000,
				FinalizedBlockHeight:  995,
				ReadyTransactionCount: 0,
			},
			metricsRepoFindByIDError:      nil,
			metricsRepoFindByIDNumOfCalls: 1,
			// MetricsRepo.GetLatestBlockMetrics
			metricsRepoGetLatestBlockMetricsReturn: &models.LatestBlockMetrics{
				BestBlockHeight:      1021,
				FinalizedBlockHeight: 1017,
			},
			metricsRepoGetLatestBlockMetricsError:      nil,
			metricsRepoGetLatestBlockMetricsNumOfCalls: 1,
			// MetricsRepo.Save
			metricsRepoSaveError:      nil,
			metricsRepoSaveNumOfCalls: 1,
		},
		{
			name:           "Invalid metrics save request",
			metricsRequest: struct{ PeerCount string }{PeerCount: "10"},
			nodeId:         "1",
			httpStatus:     http.StatusBadRequest,
		},
		{
			name: "Database fails on saving metrics",
			metricsRequest: MetricsRequest{
				PeerCount:             10,
				BestBlockHeight:       100,
				FinalizedBlockHeight:  100,
				ReadyTransactionCount: 10,
			},
			nodeId:                    "1",
			httpStatus:                http.StatusInternalServerError,
			metricsRepoSaveNumOfCalls: 1,
			metricsRepoSaveError:      errors.New("db error"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			timestamp := time.Now()

			// create mock controller
			pingRepoMock := mocks.PingRepository{}
			recordRepoMock := mocks.RecordRepository{}

			nodeRepoMock := mocks.NodeRepository{}
			nodeRepoMock.On("IsNodeOnCooldown", test.nodeId).Return(
				test.nodeRepoIsNodeOnCooldownReturns, test.nodeRepoIsNodeOnCooldownError,
			)
			nodeRepoMock.On("AddNodeToActive", test.nodeId).Return(
				test.nodeRepoAddNodeToActiveError,
			)

			metricsRepoMock := mocks.MetricsRepository{}
			metricsRepoMock.On("FindByID", test.nodeId).Return(
				test.metricsRepoFindByIDReturn, test.metricsRepoFindByIDError,
			)
			metricsRepoMock.On("GetLatestBlockMetrics").Return(
				test.metricsRepoGetLatestBlockMetricsReturn, test.metricsRepoGetLatestBlockMetricsError,
			)
			metricsRepoMock.On("Save", mock.Anything).Return(
				test.metricsRepoSaveError,
			)

			apiController := NewApiController(false, repositories.Repos{
				NodeRepo:    &nodeRepoMock,
				PingRepo:    &pingRepoMock,
				MetricsRepo: &metricsRepoMock,
				RecordRepo:  &recordRepoMock,
			}, nil)

			handler := http.HandlerFunc(apiController.SaveMetricsHandler)

			// create test request
			rb, _ := json.Marshal(test.metricsRequest)
			req, _ := http.NewRequest("POST", "/api/v1/metrics", bytes.NewReader(rb))
			c := &auth.RequestContext{
				NodeId:    test.nodeId,
				Timestamp: timestamp,
			}
			ctx := context.WithValue(req.Context(), auth.RequestContextKey, c)
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			// invoke test request
			handler.ServeHTTP(rr, req)

			// asserts
			assert.Equal(t, test.httpStatus, rr.Code, fmt.Sprintf("Response status code should be %d", test.httpStatus))

			nodeRepoMock.AssertNumberOfCalls(t, "IsNodeOnCooldown", test.nodeRepoIsNodeOnNumOfCalls)
			nodeRepoMock.AssertNumberOfCalls(t, "AddNodeToActive", test.nodeRepoAddNodeToActiveNumOfCalls)
			metricsRepoMock.AssertNumberOfCalls(t, "FindByID", test.metricsRepoFindByIDNumOfCalls)
			metricsRepoMock.AssertNumberOfCalls(t, "GetLatestBlockMetrics", test.metricsRepoGetLatestBlockMetricsNumOfCalls)
			metricsRepoMock.AssertNumberOfCalls(t, "Save", test.metricsRepoSaveNumOfCalls)
		})
	}
}
