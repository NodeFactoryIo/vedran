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
		name                  string
		metricsRequest        interface{}
		nodeId                string
		httpStatus            int
		// NodeRepo.FindByID
		nodeRepoFindByIDReturns *models.Node
		nodeRepoFindByIDError   error
		nodeRepoFindByIDNumOfCalls int
		// NodeRepo.AddNodeToActive
		nodeRepoAddNodeToActiveError error
		nodeRepoAddNodeToActiveNumOfCalls int
		// MetricsRepo.FindByID
		metricsRepoFindByIDError error
		metricsRepoFindByIDNumOfCalls int
		// MetricsRepo.Save
		metricsRepoSaveError error
		metricsRepoSaveNumOfCalls int
	}{
		{
			name: "Valid metrics save request, older metrics entry already saved",
			metricsRequest: MetricsRequest{
				PeerCount:             10,
				BestBlockHeight:       100,
				FinalizedBlockHeight:  100,
				ReadyTransactionCount: 10,
			},
			nodeId:                            "1",
			httpStatus:                        http.StatusOK,
			nodeRepoFindByIDReturns:           &models.Node{
				ID: "1",
			},
			nodeRepoFindByIDError:             nil,
			nodeRepoFindByIDNumOfCalls:        1,
			nodeRepoAddNodeToActiveError:      nil,
			nodeRepoAddNodeToActiveNumOfCalls: 0,
			metricsRepoFindByIDError:          nil,
			metricsRepoFindByIDNumOfCalls:     1,
			metricsRepoSaveError:              nil,
			metricsRepoSaveNumOfCalls:         1,
		},
		{
			name: "Valid metrics save request, first metrics entry, node should be added to active",
			metricsRequest: MetricsRequest{
				PeerCount:             10,
				BestBlockHeight:       100,
				FinalizedBlockHeight:  100,
				ReadyTransactionCount: 10,
			},
			nodeId:                            "1",
			httpStatus:                        http.StatusOK,
			nodeRepoFindByIDReturns:           &models.Node{
				ID: "1",
			},
			nodeRepoFindByIDError:             nil,
			nodeRepoFindByIDNumOfCalls:        1,
			nodeRepoAddNodeToActiveError:      nil,
			nodeRepoAddNodeToActiveNumOfCalls: 1,
			metricsRepoFindByIDError:          errors.New("not found"),
			metricsRepoFindByIDNumOfCalls:     1,
			metricsRepoSaveError:              nil,
			metricsRepoSaveNumOfCalls:         1,
		},
		{
			name:                              "Invalid metrics save request",
			metricsRequest:                    struct{ PeerCount string }{PeerCount: "10"},
			nodeId:                            "1",
			httpStatus:                        http.StatusBadRequest,
		},
		{
			name: "Database fails on fetching node",
			metricsRequest: MetricsRequest{
				PeerCount:             10,
				BestBlockHeight:       100,
				FinalizedBlockHeight:  100,
				ReadyTransactionCount: 10,
			},
			nodeId: "1",
			httpStatus:            http.StatusInternalServerError,
			nodeRepoFindByIDError:             errors.New("db error"),
			nodeRepoFindByIDNumOfCalls:        1,
		},
		{
			name: "Database fails on fetching metrics",
			metricsRequest: MetricsRequest{
				PeerCount:             10,
				BestBlockHeight:       100,
				FinalizedBlockHeight:  100,
				ReadyTransactionCount: 10,
			},
			nodeId: "1",
			httpStatus:            http.StatusInternalServerError,
			nodeRepoFindByIDError:             nil,
			nodeRepoFindByIDReturns: &models.Node{
				ID: "1",
			},
			nodeRepoFindByIDNumOfCalls:        1,
			metricsRepoFindByIDError:          errors.New("db error"),
			metricsRepoFindByIDNumOfCalls:     1,
		},
		{
			name: "Database fails on saving metrics",
			metricsRequest: MetricsRequest{
				PeerCount:             10,
				BestBlockHeight:       100,
				FinalizedBlockHeight:  100,
				ReadyTransactionCount: 10,
			},
			nodeId: "1",
			httpStatus:            http.StatusInternalServerError,
			nodeRepoFindByIDError:             nil,
			nodeRepoFindByIDReturns: &models.Node{
				ID: "1",
			},
			nodeRepoFindByIDNumOfCalls:        1,
			metricsRepoFindByIDError:          nil,
			metricsRepoFindByIDNumOfCalls:     1,
			metricsRepoSaveError: errors.New("db error"),
			metricsRepoSaveNumOfCalls: 1,
		},
		{
			name: "Adding node to active fails",
			metricsRequest: MetricsRequest{
				PeerCount:             10,
				BestBlockHeight:       100,
				FinalizedBlockHeight:  100,
				ReadyTransactionCount: 10,
			},
			nodeId: "1",
			httpStatus:            http.StatusInternalServerError,
			nodeRepoFindByIDError:             nil,
			nodeRepoFindByIDReturns: &models.Node{
				ID: "1",
			},
			nodeRepoFindByIDNumOfCalls:        1,
			metricsRepoFindByIDError:          errors.New("not found"),
			metricsRepoFindByIDNumOfCalls:     1,
			metricsRepoSaveError: nil,
			metricsRepoSaveNumOfCalls: 1,
			nodeRepoAddNodeToActiveError: errors.New("fail to add repo"),
			nodeRepoAddNodeToActiveNumOfCalls: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			timestamp := time.Now()

			// create mock controller
			pingRepoMock := mocks.PingRepository{}
			recordRepoMock := mocks.RecordRepository{}
			
			nodeRepoMock := mocks.NodeRepository{}
			nodeRepoMock.On("FindByID", test.nodeId).Return(
				test.nodeRepoFindByIDReturns, test.nodeRepoFindByIDError, 
			)
			nodeRepoMock.On("AddNodeToActive", mock.Anything).Return(
				test.nodeRepoAddNodeToActiveError,
			)
			
			metricsRepoMock := mocks.MetricsRepository{}
			// always return first value as nil, this value is not used in tested handler
			metricsRepoMock.On("FindByID", test.nodeId).Return(
				nil, test.metricsRepoFindByIDError,
			)
			metricsRepoMock.On("Save", &models.Metrics{
				NodeId:                test.nodeId,
				PeerCount:             10,
				BestBlockHeight:       100,
				FinalizedBlockHeight:  100,
				ReadyTransactionCount: 10,
			}).Return(test.metricsRepoSaveError)
			
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
			
			nodeRepoMock.AssertNumberOfCalls(t, "FindByID", test.nodeRepoFindByIDNumOfCalls)
			nodeRepoMock.AssertNumberOfCalls(t, "AddNodeToActive", test.nodeRepoAddNodeToActiveNumOfCalls)
			metricsRepoMock.AssertNumberOfCalls(t, "Save", test.metricsRepoSaveNumOfCalls)
			metricsRepoMock.AssertNumberOfCalls(t, "FindByID", test.metricsRepoFindByIDNumOfCalls)
		})
	}
}
