package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/NodeFactoryIo/vedran/internal/configuration"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"github.com/NodeFactoryIo/vedran/internal/rpc"
	actionMocks "github.com/NodeFactoryIo/vedran/mocks/actions"
	tunnelMocks "github.com/NodeFactoryIo/vedran/mocks/http-tunnel/server"
	repoMocks "github.com/NodeFactoryIo/vedran/mocks/repositories"
	"github.com/stretchr/testify/mock"
)

var (
	mux    *http.ServeMux
	server *httptest.Server
)

type handleFnMock func(http.ResponseWriter, *http.Request)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
}

func teardown() {
	server.Close()
}

func TestApiController_RPCHandler(t *testing.T) {
	setup()
	defer teardown()

	poolerMock := &tunnelMocks.Pooler{}
	configuration.Config.PortPool = poolerMock

	nodeRepoMock := repoMocks.NodeRepository{}
	nodeRepoMock.On("UpdateNodeUsed", mock.Anything).Return()
	recordRepoMock := repoMocks.RecordRepository{}
	recordRepoMock.On("Save", mock.Anything).Return(nil)

	actionsMockObject := new(actionMocks.Actions)
	actionsMockObject.On("PenalizeNode", mock.Anything, mock.Anything).Return()

	apiController := NewApiController(false, repositories.Repos{
		NodeRepo:   &nodeRepoMock,
		RecordRepo: &recordRepoMock,
	}, actionsMockObject, "")

	handler := http.HandlerFunc(apiController.RPCHandler)

	tests := []struct {
		name        string
		rpcRequest  string
		rpcResponse rpc.RPCResponse
		nodes       []models.Node
		handleFunc  handleFnMock
	}{
		{
			name:       "Returns response if node returnes valid rpc response",
			rpcRequest: `{"jsonrpc": "2.0", "id": 1, "method": "system"}`,
			rpcResponse: rpc.RPCResponse{
				ID:      1,
				JSONRPC: "2.0",
				Error:   nil,
			},
			nodes: []models.Node{{ID: "test-id"}},
			handleFunc: func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, `{"id": 1, "jsonrpc": "2.0"}`)
			}},
		{
			name:       "Returns parse error if json invalid",
			rpcRequest: `INVALID`,
			rpcResponse: rpc.RPCResponse{
				ID:      0,
				JSONRPC: "2.0",
				Error:   &rpc.RPCError{Code: -32700, Message: "Parse error"},
			},
			handleFunc: func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, `{"id": 1, "jsonrpc": "2.0"}`)
			}},
		{
			name:       "Returns server error if no available nodes",
			rpcRequest: `{"jsonrpc": "2.0", "id": 1, "method": "system"}`,
			rpcResponse: rpc.RPCResponse{
				ID:      1,
				JSONRPC: "2.0",
				Error:   &rpc.RPCError{Code: -32603, Message: "No available nodes"},
			},
			nodes: []models.Node{},
			handleFunc: func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, `{"id": 1, "jsonrpc": "2.0"}`)
			}},
		{
			name:       "Returns server error if all nodes return invalid rpc response",
			rpcRequest: `{"jsonrpc": "2.0", "id": 1, "method": "system"}`,
			rpcResponse: rpc.RPCResponse{
				ID:      1,
				JSONRPC: "2.0",
				Error:   &rpc.RPCError{Code: -32603, Message: "Internal Server Error"},
			},
			nodes: []models.Node{{ID: "test-id"}},
			handleFunc: func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, `{"id": 1, "jsonrpc": "2.0"}`)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setup()

			serverURL, _ := url.Parse(server.URL)
			port, _ := strconv.Atoi(serverURL.Port())
			poolerMock.On("GetHTTPPort", mock.Anything).Once().Return(port, nil)

			mux.HandleFunc("/", test.handleFunc)

			nodeRepoMock.On("GetActiveNodes", mock.Anything).Return(&test.nodes, nil)

			req, _ := http.NewRequest("POST", "/", bytes.NewReader([]byte(test.rpcRequest)))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			var body rpc.RPCResponse
			_ = json.Unmarshal(rr.Body.Bytes(), &body)

			if test.rpcResponse != (rpc.RPCResponse{}) && !reflect.DeepEqual(body, test.rpcResponse) {
				t.Errorf("SendRequestToNode() body = %v, want %v", body, test.rpcResponse)
				return
			}

			teardown()
		})
	}
}

func TestApiController_BatchRPCHandler(t *testing.T) {
	setup()
	defer teardown()

	poolerMock := &tunnelMocks.Pooler{}
	configuration.Config.PortPool = poolerMock

	nodeRepoMock := repoMocks.NodeRepository{}
	nodeRepoMock.On("UpdateNodeUsed", mock.Anything).Return()
	recordRepoMock := repoMocks.RecordRepository{}
	recordRepoMock.On("Save", mock.Anything).Return(nil)

	actionsMockObject := new(actionMocks.Actions)
	actionsMockObject.On("PenalizeNode", mock.Anything, mock.Anything).Return()

	apiController := NewApiController(false, repositories.Repos{
		NodeRepo:   &nodeRepoMock,
		RecordRepo: &recordRepoMock,
	}, actionsMockObject, "")

	handler := http.HandlerFunc(apiController.RPCHandler)

	tests := []struct {
		name         string
		rpcRequest   string
		rpcResponses []rpc.RPCResponse
		nodes        []models.Node
		handleFunc   handleFnMock
	}{
		{
			name:       "Returns batch server error if all nodes return invalid rpc response",
			rpcRequest: `[{"jsonrpc": "2.0", "id": 1, "method": "system"}]`,
			rpcResponses: []rpc.RPCResponse{
				{
					ID:      1,
					JSONRPC: "2.0",
					Error:   &rpc.RPCError{Code: -32603, Message: "Internal Server Error"}},
			},
			nodes: []models.Node{{ID: "test-id"}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setup()

			serverURL, _ := url.Parse(server.URL)
			port, _ := strconv.Atoi(serverURL.Port())
			poolerMock.On("GetHTTPPort", mock.Anything).Once().Return(port, nil)

			if test.handleFunc != nil {
				mux.HandleFunc("/", test.handleFunc)
			}

			nodeRepoMock.On("GetActiveNodes", mock.Anything).Return(&test.nodes, nil)

			req, _ := http.NewRequest("POST", "/", bytes.NewReader([]byte(test.rpcRequest)))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			var body []rpc.RPCResponse
			_ = json.Unmarshal(rr.Body.Bytes(), &body)

			if !reflect.DeepEqual(body, test.rpcResponses) {
				t.Errorf("SendRequestToNode() body = %v, want %v", body, test.rpcResponses)
				return
			}

			teardown()
		})
	}
}

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}

func TestApiController_RPCHandler_InvalidBody(t *testing.T) {
	setup()
	defer teardown()
	tests := []struct {
		name        string
		rpcResponse rpc.RPCResponse
	}{
		{
			name: "Returns parse error if reading request body fails",
			rpcResponse: rpc.RPCResponse{
				ID:      0,
				JSONRPC: "2.0",
				Error:   &rpc.RPCError{Code: -32700, Message: "Parse error"}}},
	}

	actionsMockObject := new(actionMocks.Actions)
	actionsMockObject.On("PenalizeNode", mock.Anything, mock.Anything).Return()

	apiController := NewApiController(false, repositories.Repos{}, actionsMockObject, "")

	handler := http.HandlerFunc(apiController.RPCHandler)

	for _, test := range tests {
		setup()

		req, _ := http.NewRequest("POST", "/", errReader(0))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		var body rpc.RPCResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &body)

		if !reflect.DeepEqual(body, test.rpcResponse) {
			t.Errorf("SendRequestToNode() body = %v, want %v", body, test.rpcResponse)
			return
		}

		teardown()
	}
}
