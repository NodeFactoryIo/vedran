package controllers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	mm "github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/NodeFactoryIo/vedran/internal/configuration"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	actionMocks "github.com/NodeFactoryIo/vedran/mocks/actions"
	tunnelMocks "github.com/NodeFactoryIo/vedran/mocks/http-tunnel/server"
	mocks "github.com/NodeFactoryIo/vedran/mocks/repositories"
)

func TestApiController_WSHandler(t *testing.T) {
	tests := []struct {
		name                         string
		nodeRepoGetActiveNodesReturn []models.Node
		penalizeNodeNumOfCalls       int
		updateNodeUsedNumOfCalls     int
		requestType                  string
		expectedResponses            []string
		forceNodeWStoFail            bool
	}{
		{
			name: "test simple request",
			nodeRepoGetActiveNodesReturn: []models.Node{
				{ID: "1", ConfigHash: "", PayoutAddress: "", Token: "", Cooldown: 0, LastUsed: 1, Active: true},
			},
			// 1 for starting connection + 1 for successful response
			updateNodeUsedNumOfCalls: 2,
			requestType:              SimpleRequest,
			expectedResponses:        []string{SimpleRequest},
		},
		{
			name: "test subscription request",
			nodeRepoGetActiveNodesReturn: []models.Node{
				{ID: "1", ConfigHash: "", PayoutAddress: "", Token: "", Cooldown: 0, LastUsed: 1, Active: true},
			},
			// 1 for starting connection + 5 for successful responses
			updateNodeUsedNumOfCalls: 6,
			requestType:              SubscribeRequest,
			expectedResponses: []string{
				"subscription response 1",
				"subscription response 2",
				"subscription response 3",
				"subscription response 4",
				"subscription response 5",
			},
		},
		{
			name: "test failed request",
			nodeRepoGetActiveNodesReturn: []models.Node{
				{ID: "1", ConfigHash: "", PayoutAddress: "", Token: "", Cooldown: 0, LastUsed: 1, Active: true},
			},
			requestType:            SimpleRequest,
			expectedResponses:      []string{},
			forceNodeWStoFail:      true,
			penalizeNodeNumOfCalls: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			nodeRepoMock := mocks.NodeRepository{}
			nodeRepoMock.On("GetActiveNodes", mock.Anything).Return(&test.nodeRepoGetActiveNodesReturn)
			nodeRepoMock.On("UpdateNodeUsed", mock.Anything).Return()

			recordRepoMock := mocks.RecordRepository{}
			recordRepoMock.On("Save", mock.Anything).Return(nil)

			actionsMockObject := new(actionMocks.Actions)
			actionsMockObject.On(
				"PenalizeNode",
				mock.MatchedBy(func(n models.Node) bool { return n.ID == "1" }),
				mock.Anything,
				mock.Anything,
			).Return()

			apiController := NewApiController(false, repositories.Repos{
				NodeRepo:   &nodeRepoMock,
				RecordRepo: &recordRepoMock,
			}, actionsMockObject)

			// start test loadbalancer ws server
			router := mm.NewRouter()
			router.HandleFunc("/ws", apiController.WSHandler)
			s := httptest.NewServer(router)
			defer s.Close()

			// start mock node ws server
			ns := StartMockNodeWS(test.forceNodeWStoFail)
			defer ns.Close()

			// mock pooler
			strPort := strings.Split(strings.Split(ns.URL, ":")[2], "/")[0]
			nodePort, _ := strconv.Atoi(strPort)
			poolerMock := tunnelMocks.Pooler{}
			poolerMock.On("GetWSPort", mock.Anything).Return(nodePort, nil)
			configuration.Config.PortPool = &poolerMock

			// ws dial loadbalancer
			u := "ws" + strings.TrimPrefix(s.URL, "http") + "/ws"
			ws, _, err := websocket.DefaultDialer.Dial(u, nil)
			assert.NoError(t, err)
			defer ws.Close()

			go func() {
				for _, respons := range test.expectedResponses {
					_, msg, err2 := ws.ReadMessage()
					assert.NoError(t, err2)
					assert.Equal(t, respons, string(msg))
				}
			}()

			_ = ws.WriteMessage(1, []byte(test.requestType))
			time.Sleep(1 * time.Second)
			actionsMockObject.AssertNumberOfCalls(t, "PenalizeNode", test.penalizeNodeNumOfCalls)
			nodeRepoMock.AssertNumberOfCalls(t, "UpdateNodeUsed", test.updateNodeUsedNumOfCalls)

			// cleanup
			configuration.Config.PortPool = nil
		})
	}
}

func StartMockNodeWS(shouldFail bool) *httptest.Server {
	mockNodeWs := MockNodeWs{shouldFail: shouldFail}
	nodeRouter := mm.NewRouter()
	nodeRouter.HandleFunc("/", mockNodeWs.MockEchoWsHandler)
	return httptest.NewServer(nodeRouter)
}

const (
	SimpleRequest    = "request"
	SubscribeRequest = "subscribe"
	FailRequest      = "fail"
)

type MockNodeWs struct {
	shouldFail bool
}

func (n *MockNodeWs) MockEchoWsHandler(w http.ResponseWriter, r *http.Request) {
	if n.shouldFail {
		http.Error(w, "Failed establishing connection", 500)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed upgrading connection", 500)
		return
	}
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		switch string(msg) {
		case SimpleRequest:
			err = conn.WriteMessage(msgType, msg)
			if err != nil {
				return
			}
		case SubscribeRequest:
			// emulate subscription behaviour
			for i := 0; i < 5; i++ {
				err = conn.WriteMessage(msgType, []byte(fmt.Sprintf("subscription response %d", i+1)))
				if err != nil {
					return
				}
				time.Sleep(1 * time.Microsecond)
			}
		case FailRequest:
			err := conn.Close()
			if err != nil {
				return
			}
		default:
			return
		}
	}
}
