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
		// MetricsRepo.Save
		metricsRepoSaveAndCheckIfFirstEntryError      error
		metricsRepoSaveAndCheckIfFirstEntryReturn     bool
		metricsRepoSaveAndCheckIfFirstEntryNumOfCalls int
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
			nodeRepoFindByIDError:                         nil,
			nodeRepoFindByIDNumOfCalls:                    0,
			nodeRepoAddNodeToActiveError:                  nil,
			nodeRepoAddNodeToActiveNumOfCalls:             0,
			metricsRepoSaveAndCheckIfFirstEntryNumOfCalls: 1,
			metricsRepoSaveAndCheckIfFirstEntryError:      nil,
			metricsRepoSaveAndCheckIfFirstEntryReturn:     false,
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
			metricsRepoSaveAndCheckIfFirstEntryNumOfCalls: 1,
			metricsRepoSaveAndCheckIfFirstEntryError:      nil,
			metricsRepoSaveAndCheckIfFirstEntryReturn:     true,
		},
		{
			name:                              "Invalid metrics save request",
			metricsRequest:                    struct{ PeerCount string }{PeerCount: "10"},
			nodeId:                            "1",
			httpStatus:                        http.StatusBadRequest,
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
			metricsRepoSaveAndCheckIfFirstEntryNumOfCalls: 1,
			metricsRepoSaveAndCheckIfFirstEntryError:      errors.New("db error"),
			metricsRepoSaveAndCheckIfFirstEntryReturn:     false,
		},
		{
			name: "Database fails on fetching node, first metrics entry",
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
			metricsRepoSaveAndCheckIfFirstEntryNumOfCalls: 1,
			metricsRepoSaveAndCheckIfFirstEntryError:      nil,
			metricsRepoSaveAndCheckIfFirstEntryReturn:     true,
		},
		{
			name: "Adding node to active fails, first metrics entry",
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
			nodeRepoAddNodeToActiveError: errors.New("fail to add repo"),
			nodeRepoAddNodeToActiveNumOfCalls: 1,
			metricsRepoSaveAndCheckIfFirstEntryNumOfCalls: 1,
			metricsRepoSaveAndCheckIfFirstEntryError:      nil,
			metricsRepoSaveAndCheckIfFirstEntryReturn:     true,
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

			metricsRepoMock.On("SaveAndCheckIfFirstEntry", &models.Metrics{
				NodeId:                test.nodeId,
				PeerCount:             10,
				BestBlockHeight:       100,
				FinalizedBlockHeight:  100,
				ReadyTransactionCount: 10,
			}).Return(test.metricsRepoSaveAndCheckIfFirstEntryReturn, test.metricsRepoSaveAndCheckIfFirstEntryError)
			
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
			metricsRepoMock.AssertNumberOfCalls(t, "SaveAndCheckIfFirstEntry", test.metricsRepoSaveAndCheckIfFirstEntryNumOfCalls)
		})
	}
}
