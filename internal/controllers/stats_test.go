package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NodeFactoryIo/go-substrate-rpc-client/signature"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	mocks "github.com/NodeFactoryIo/vedran/mocks/repositories"
	"github.com/ethereum/go-ethereum/common/hexutil"
	muxhelpper "github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestApiController_StatisticsHandlerAllStats(t *testing.T) {
	now := time.Now()
	getNow = func() time.Time {
		return now
	}
	tests := []struct {
		name          string
		httpStatus    int
		nodeId        string
		payoutAddress string
		// NodeRepo.GetAll
		nodeRepoGetAllReturns *[]models.Node
		nodeRepoGetAllError   error
		// RecordRepo.FindSuccessfulRecordsInsideInterval
		recordRepoFindSuccessfulRecordsInsideIntervalReturns []models.Record
		recordRepoFindSuccessfulRecordsInsideIntervalError   error
		// DowntimeRepo.FindDowntimesInsideInterval
		downtimeRepoFindDowntimesInsideIntervalReturns []models.Downtime
		downtimeRepoFindDowntimesInsideIntervalError   error
		// PingRepo.CalculateDowntime
		pingRepoCalculateDowntimeReturnDuration time.Duration
		pingRepoCalculateDowntimeError          error
		// PayoutRepo.FindLatestPayout
		payoutRepoFindLatestPayoutReturns *models.Payout
		payoutRepoFindLatestPayoutError   error
		// Stats
		nodeNumberOfPings    float64
		nodeNumberOfRequests float64
	}{
		{
			name:          "get valid stats",
			nodeId:        "1",
			payoutAddress: "0xtest-address",
			httpStatus:    http.StatusOK,
			// NodeRepo.GetAll
			nodeRepoGetAllReturns: &[]models.Node{
				{
					ID:            "1",
					PayoutAddress: "0xtest-address",
				},
			},
			nodeRepoGetAllError: nil,
			// RecordRepo.FindSuccessfulRecordsInsideInterval
			recordRepoFindSuccessfulRecordsInsideIntervalReturns: nil,
			recordRepoFindSuccessfulRecordsInsideIntervalError:   errors.New("not found"),
			// DowntimeRepo.FindDowntimesInsideInterval
			downtimeRepoFindDowntimesInsideIntervalReturns: nil,
			downtimeRepoFindDowntimesInsideIntervalError:   errors.New("not found"),
			// PingRepo.CalculateDowntime
			pingRepoCalculateDowntimeReturnDuration: 5 * time.Second,
			pingRepoCalculateDowntimeError:          nil,
			// PayoutRepo.FindLatestPayout
			payoutRepoFindLatestPayoutReturns: &models.Payout{
				Timestamp:      now.Add(-24 * time.Hour),
				PaymentDetails: nil,
			},
			payoutRepoFindLatestPayoutError: nil,
			// Stats
			nodeNumberOfRequests: float64(0),
			nodeNumberOfPings:    float64(8640),
		},
		{
			name:                            "unable to get latest interval, server error",
			httpStatus:                      http.StatusInternalServerError,
			payoutRepoFindLatestPayoutError: errors.New("db-error"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create mock controller
			nodeRepoMock := mocks.NodeRepository{}
			nodeRepoMock.On("GetAll").Return(
				test.nodeRepoGetAllReturns, test.nodeRepoGetAllError,
			)
			nodeRepoMock.On("FindByID", test.nodeId).Return(&models.Node{
				ID:            test.nodeId,
				PayoutAddress: test.payoutAddress,
			}, nil)
			recordRepoMock := mocks.RecordRepository{}
			recordRepoMock.On("FindSuccessfulRecordsInsideInterval",
				test.nodeId, mock.Anything, mock.Anything,
			).Return(
				test.recordRepoFindSuccessfulRecordsInsideIntervalReturns,
				test.recordRepoFindSuccessfulRecordsInsideIntervalError,
			)
			metricsRepoMock := mocks.MetricsRepository{}
			pingRepoMock := mocks.PingRepository{}
			pingRepoMock.On("CalculateDowntime",
				test.nodeId, mock.Anything,
			).Return(
				time.Now(),
				test.pingRepoCalculateDowntimeReturnDuration,
				test.pingRepoCalculateDowntimeError,
			)
			downtimeRepoMock := mocks.DowntimeRepository{}
			downtimeRepoMock.On("FindDowntimesInsideInterval",
				test.nodeId, mock.Anything, mock.Anything,
			).Return(
				test.downtimeRepoFindDowntimesInsideIntervalReturns,
				test.downtimeRepoFindDowntimesInsideIntervalError,
			)
			payoutRepoMock := mocks.PayoutRepository{}
			payoutRepoMock.On("FindLatestPayout").Return(
				test.payoutRepoFindLatestPayoutReturns,
				test.payoutRepoFindLatestPayoutError,
			)
			payoutRepoMock.On("Save", mock.Anything).Return(nil)
			apiController := NewApiController(false, repositories.Repos{
				NodeRepo:     &nodeRepoMock,
				PingRepo:     &pingRepoMock,
				MetricsRepo:  &metricsRepoMock,
				RecordRepo:   &recordRepoMock,
				DowntimeRepo: &downtimeRepoMock,
				PayoutRepo:   &payoutRepoMock,
			}, nil, "")
			handler := http.HandlerFunc(apiController.StatisticsHandlerAllStats)
			req, _ := http.NewRequest("GET", "/api/v1/stats", bytes.NewReader(nil))
			rr := httptest.NewRecorder()

			// invoke test request
			handler.ServeHTTP(rr, req)

			// asserts
			assert.Equal(t, test.httpStatus, rr.Code, fmt.Sprintf("Response status code should be %d", test.httpStatus))

			var statsResponse StatsResponse
			if rr.Code == http.StatusOK {
				_ = json.Unmarshal(rr.Body.Bytes(), &statsResponse)
				assert.LessOrEqual(t, test.nodeNumberOfPings, statsResponse.Stats[test.payoutAddress].TotalPings)
				assert.Equal(t, test.nodeNumberOfRequests, statsResponse.Stats[test.payoutAddress].TotalRequests)
			}
		})
	}
}

func TestApiController_StatisticsHandlerAllStatsForLoadbalancer(t *testing.T) {
	now := time.Now()
	getNow = func() time.Time {
		return now
	}
	tests := []struct {
		name          string
		httpStatus    int
		nodeId        string
		payoutAddress string
		// NodeRepo.GetAll
		nodeRepoGetAllReturns *[]models.Node
		nodeRepoGetAllError   error
		// RecordRepo.FindSuccessfulRecordsInsideInterval
		recordRepoFindSuccessfulRecordsInsideIntervalReturns []models.Record
		recordRepoFindSuccessfulRecordsInsideIntervalError   error
		// DowntimeRepo.FindDowntimesInsideInterval
		downtimeRepoFindDowntimesInsideIntervalReturns []models.Downtime
		downtimeRepoFindDowntimesInsideIntervalError   error
		// PingRepo.CalculateDowntime
		pingRepoCalculateDowntimeReturnDuration time.Duration
		pingRepoCalculateDowntimeError          error
		// PayoutRepo.FindLatestPayout
		payoutRepoFindLatestPayoutReturns *models.Payout
		payoutRepoFindLatestPayoutError   error
		// Stats
		nodeNumberOfPings    float64
		nodeNumberOfRequests float64
		//
		requestContent string
		//
		secret        string
		signatureData string
	}{
		{
			name:          "get valid stats, 200 OK",
			nodeId:        "1",
			payoutAddress: "0xtest-address",
			httpStatus:    http.StatusOK,
			// NodeRepo.GetAll
			nodeRepoGetAllReturns: &[]models.Node{
				{
					ID:            "1",
					PayoutAddress: "0xtest-address",
				},
			},
			nodeRepoGetAllError: nil,
			// RecordRepo.FindSuccessfulRecordsInsideInterval
			recordRepoFindSuccessfulRecordsInsideIntervalReturns: nil,
			recordRepoFindSuccessfulRecordsInsideIntervalError:   errors.New("not found"),
			// DowntimeRepo.FindDowntimesInsideInterval
			downtimeRepoFindDowntimesInsideIntervalReturns: nil,
			downtimeRepoFindDowntimesInsideIntervalError:   errors.New("not found"),
			// PingRepo.CalculateDowntime
			pingRepoCalculateDowntimeReturnDuration: 5 * time.Second,
			pingRepoCalculateDowntimeError:          nil,
			// PayoutRepo.FindLatestPayout
			payoutRepoFindLatestPayoutReturns: &models.Payout{
				Timestamp:      now.Add(-24 * time.Hour),
				PaymentDetails: nil,
			},
			payoutRepoFindLatestPayoutError: nil,
			// Stats
			nodeNumberOfRequests: float64(0),
			nodeNumberOfPings:    float64(8640),
			//
			requestContent: `{"total_reward":"1000000"}`,
			//
			secret:        "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			signatureData: StatsSignedData,
		},
		{
			name:          "missing signature, 400 bad request",
			nodeId:        "1",
			payoutAddress: "0xtest-address",
			httpStatus:    http.StatusBadRequest,
			// NodeRepo.GetAll
			nodeRepoGetAllReturns: &[]models.Node{
				{
					ID:            "1",
					PayoutAddress: "0xtest-address",
				},
			},
			nodeRepoGetAllError: nil,
			// RecordRepo.FindSuccessfulRecordsInsideInterval
			recordRepoFindSuccessfulRecordsInsideIntervalReturns: nil,
			recordRepoFindSuccessfulRecordsInsideIntervalError:   errors.New("not found"),
			// DowntimeRepo.FindDowntimesInsideInterval
			downtimeRepoFindDowntimesInsideIntervalReturns: nil,
			downtimeRepoFindDowntimesInsideIntervalError:   errors.New("not found"),
			// PingRepo.CalculateDowntime
			pingRepoCalculateDowntimeReturnDuration: 5 * time.Second,
			pingRepoCalculateDowntimeError:          nil,
			// PayoutRepo.FindLatestPayout
			payoutRepoFindLatestPayoutReturns: &models.Payout{
				Timestamp:      now.Add(-24 * time.Hour),
				PaymentDetails: nil,
			},
			payoutRepoFindLatestPayoutError: nil,
			// Stats
			nodeNumberOfRequests: float64(0),
			nodeNumberOfPings:    float64(8640),
			//
			secret:        "",
			signatureData: StatsSignedData,
		},
		{
			name:          "invalid signature, 400 bad request",
			nodeId:        "1",
			payoutAddress: "0xtest-address",
			httpStatus:    http.StatusBadRequest,
			// NodeRepo.GetAll
			nodeRepoGetAllReturns: &[]models.Node{
				{
					ID:            "1",
					PayoutAddress: "0xtest-address",
				},
			},
			nodeRepoGetAllError: nil,
			// RecordRepo.FindSuccessfulRecordsInsideInterval
			recordRepoFindSuccessfulRecordsInsideIntervalReturns: nil,
			recordRepoFindSuccessfulRecordsInsideIntervalError:   errors.New("not found"),
			// DowntimeRepo.FindDowntimesInsideInterval
			downtimeRepoFindDowntimesInsideIntervalReturns: nil,
			downtimeRepoFindDowntimesInsideIntervalError:   errors.New("not found"),
			// PingRepo.CalculateDowntime
			pingRepoCalculateDowntimeReturnDuration: 5 * time.Second,
			pingRepoCalculateDowntimeError:          nil,
			// PayoutRepo.FindLatestPayout
			payoutRepoFindLatestPayoutReturns: &models.Payout{
				Timestamp:      now.Add(-24 * time.Hour),
				PaymentDetails: nil,
			},
			payoutRepoFindLatestPayoutError: nil,
			// Stats
			nodeNumberOfRequests: float64(0),
			nodeNumberOfPings:    float64(8640),
			//
			secret:        "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			signatureData: "loadbalancer-invalid-request",
		},
		{
			name:                            "unable to get latest interval, 500 server error",
			httpStatus:                      http.StatusInternalServerError,
			payoutRepoFindLatestPayoutError: errors.New("db-error"),
			secret:                          "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			requestContent:                  `{"total_reward":"1000000"}`,
			signatureData:                   StatsSignedData,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create mock controller
			nodeRepoMock := mocks.NodeRepository{}
			nodeRepoMock.On("GetAll").Return(
				test.nodeRepoGetAllReturns, test.nodeRepoGetAllError,
			)
			nodeRepoMock.On("FindByID", test.nodeId).Return(&models.Node{
				ID:            test.nodeId,
				PayoutAddress: "0xtest-address",
			}, nil)
			recordRepoMock := mocks.RecordRepository{}
			recordRepoMock.On("FindSuccessfulRecordsInsideInterval",
				test.nodeId, mock.Anything, mock.Anything,
			).Return(
				test.recordRepoFindSuccessfulRecordsInsideIntervalReturns,
				test.recordRepoFindSuccessfulRecordsInsideIntervalError,
			)
			metricsRepoMock := mocks.MetricsRepository{}
			pingRepoMock := mocks.PingRepository{}
			pingRepoMock.On("CalculateDowntime",
				test.nodeId, mock.Anything,
			).Return(
				time.Now(),
				test.pingRepoCalculateDowntimeReturnDuration,
				test.pingRepoCalculateDowntimeError,
			)
			downtimeRepoMock := mocks.DowntimeRepository{}
			downtimeRepoMock.On("FindDowntimesInsideInterval",
				test.nodeId, mock.Anything, mock.Anything,
			).Return(
				test.downtimeRepoFindDowntimesInsideIntervalReturns,
				test.downtimeRepoFindDowntimesInsideIntervalError,
			)
			payoutRepoMock := mocks.PayoutRepository{}
			payoutRepoMock.On("FindLatestPayout").Return(
				test.payoutRepoFindLatestPayoutReturns,
				test.payoutRepoFindLatestPayoutError,
			)
			payoutRepoMock.On("Save", mock.Anything).Return(nil)
			apiController := NewApiController(false, repositories.Repos{
				NodeRepo:     &nodeRepoMock,
				PingRepo:     &pingRepoMock,
				MetricsRepo:  &metricsRepoMock,
				RecordRepo:   &recordRepoMock,
				DowntimeRepo: &downtimeRepoMock,
				PayoutRepo:   &payoutRepoMock,
			}, nil, test.secret)
			handler := http.HandlerFunc(apiController.StatisticsHandlerAllStatsForLoadbalancer)

			req, _ := http.NewRequest("POST", "/api/v1/stats", bytes.NewReader([]byte(test.requestContent)))

			if test.secret != "" {
				sig, _ := signature.Sign([]byte(test.signatureData), test.secret)
				req.Header.Set("X-Signature", hexutil.Encode(sig))
			}

			rr := httptest.NewRecorder()

			// invoke test request
			handler.ServeHTTP(rr, req)

			// asserts
			assert.Equal(t, test.httpStatus, rr.Code, fmt.Sprintf("Response status code should be %d", test.httpStatus))

			var statsResponse LoadbalancerStatsResponse
			if rr.Code == http.StatusOK {
				_ = json.Unmarshal(rr.Body.Bytes(), &statsResponse)
				assert.LessOrEqual(t, test.nodeNumberOfPings, statsResponse.Stats[test.payoutAddress].TotalPings)
				assert.Equal(t, test.nodeNumberOfRequests, statsResponse.Stats[test.payoutAddress].TotalRequests)
			}
		})
	}
}

func TestApiController_StatisticsHandlerStatsForNode(t *testing.T) {
	now := time.Now()
	getNow = func() time.Time {
		return now
	}
	tests := []struct {
		name       string
		httpStatus int
		nodeId     string
		contextKey string
		// RecordRepo.FindSuccessfulRecordsInsideInterval
		recordRepoFindSuccessfulRecordsInsideIntervalReturns []models.Record
		recordRepoFindSuccessfulRecordsInsideIntervalError   error
		// DowntimeRepo.FindDowntimesInsideInterval
		downtimeRepoFindDowntimesInsideIntervalReturns []models.Downtime
		downtimeRepoFindDowntimesInsideIntervalError   error
		// PingRepo.CalculateDowntime
		pingRepoCalculateDowntimeReturnDuration time.Duration
		pingRepoCalculateDowntimeError          error
		// PayoutRepo.FindLatestPayout
		payoutRepoFindLatestPayoutReturns *models.Payout
		payoutRepoFindLatestPayoutError   error
		// Stats
		nodeNumberOfPings    float64
		nodeNumberOfRequests float64
	}{
		{
			name:       "get valid stats",
			nodeId:     "1",
			httpStatus: http.StatusOK,
			contextKey: "id",
			// RecordRepo.FindSuccessfulRecordsInsideInterval
			recordRepoFindSuccessfulRecordsInsideIntervalReturns: nil,
			recordRepoFindSuccessfulRecordsInsideIntervalError:   errors.New("not found"),
			// DowntimeRepo.FindDowntimesInsideInterval
			downtimeRepoFindDowntimesInsideIntervalReturns: nil,
			downtimeRepoFindDowntimesInsideIntervalError:   errors.New("not found"),
			// PingRepo.CalculateDowntime
			pingRepoCalculateDowntimeReturnDuration: 5 * time.Second,
			pingRepoCalculateDowntimeError:          nil,
			// PayoutRepo.FindLatestPayout
			payoutRepoFindLatestPayoutReturns: &models.Payout{
				Timestamp:      now.Add(-24 * time.Hour),
				PaymentDetails: nil,
			},
			payoutRepoFindLatestPayoutError: nil,
			// Stats
			nodeNumberOfRequests: float64(0),
			nodeNumberOfPings:    float64(8640),
		},
		{
			name:                            "unable to get latest interval, server error",
			httpStatus:                      http.StatusInternalServerError,
			contextKey:                      "id",
			payoutRepoFindLatestPayoutError: errors.New("db-error"),
		},
		{
			name:                            "unable to find id from request, server error",
			httpStatus:                      http.StatusNotFound,
			contextKey:                      "id",
			payoutRepoFindLatestPayoutError: errors.New("not found"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create mock controller
			recordRepoMock := mocks.RecordRepository{}
			recordRepoMock.On("FindSuccessfulRecordsInsideInterval",
				test.nodeId, mock.Anything, mock.Anything,
			).Return(
				test.recordRepoFindSuccessfulRecordsInsideIntervalReturns,
				test.recordRepoFindSuccessfulRecordsInsideIntervalError,
			)
			metricsRepoMock := mocks.MetricsRepository{}
			pingRepoMock := mocks.PingRepository{}
			pingRepoMock.On("CalculateDowntime",
				test.nodeId, mock.Anything,
			).Return(
				time.Now(),
				test.pingRepoCalculateDowntimeReturnDuration,
				test.pingRepoCalculateDowntimeError,
			)
			downtimeRepoMock := mocks.DowntimeRepository{}
			downtimeRepoMock.On("FindDowntimesInsideInterval",
				test.nodeId, mock.Anything, mock.Anything,
			).Return(
				test.downtimeRepoFindDowntimesInsideIntervalReturns,
				test.downtimeRepoFindDowntimesInsideIntervalError,
			)
			payoutRepoMock := mocks.PayoutRepository{}
			payoutRepoMock.On("FindLatestPayout").Return(
				test.payoutRepoFindLatestPayoutReturns,
				test.payoutRepoFindLatestPayoutError,
			)
			apiController := NewApiController(false, repositories.Repos{
				PingRepo:     &pingRepoMock,
				MetricsRepo:  &metricsRepoMock,
				RecordRepo:   &recordRepoMock,
				DowntimeRepo: &downtimeRepoMock,
				PayoutRepo:   &payoutRepoMock,
			}, nil, "")
			type ContextKey string
			req, _ := http.NewRequest("GET", "/api/v1/stats/node/1", bytes.NewReader(nil))
			req = req.WithContext(context.WithValue(req.Context(), ContextKey(test.contextKey), "1"))
			rr := httptest.NewRecorder()

			// invoke test request
			router := muxhelpper.NewRouter()
			router.HandleFunc("/api/v1/stats/node/{id}", apiController.StatisticsHandlerStatsForNode)
			router.ServeHTTP(rr, req)

			// asserts
			assert.Equal(t, test.httpStatus, rr.Code, fmt.Sprintf("Response status code should be %d", test.httpStatus))

			var statsResponse models.NodeStatsDetails
			if rr.Code == http.StatusOK {
				_ = json.Unmarshal(rr.Body.Bytes(), &statsResponse)
				assert.LessOrEqual(t, test.nodeNumberOfPings, statsResponse.TotalPings)
				assert.Equal(t, test.nodeNumberOfRequests, statsResponse.TotalRequests)
			}
		})
	}
}
