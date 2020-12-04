package rpc

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/NodeFactoryIo/vedran/internal/configuration"
	"github.com/NodeFactoryIo/vedran/internal/models"
	tunnelMocks "github.com/NodeFactoryIo/vedran/mocks/http-tunnel/server"
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

func TestIsBatch(t *testing.T) {
	type args struct {
		reqBody []byte
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Returns false if request is not an array",
			args: args{[]byte(" {}")},
			want: false},
		{
			name: "Returns ture if request is an array",
			args: args{[]byte(" []")},
			want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBatch(tt.args.reqBody); got != tt.want {
				t.Errorf("IsBatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendRequestToNode(t *testing.T) {
	setup()
	defer teardown()

	poolerMock := &tunnelMocks.Pooler{}
	configuration.Config.PortPool = poolerMock

	type args struct {
		url     string
		batch   bool
		node    models.Node
		reqBody []byte
	}
	tests := []struct {
		name       string
		portValid  bool
		args       args
		want       []byte
		wantErr    bool
		handleFunc handleFnMock
	}{
		{
			name:      "Returns error if getting port fails ",
			args:      args{"valid", true, models.Node{}, []byte("{}")},
			portValid: false,
			wantErr:   true,
			want:      nil,
			handleFunc: func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, `{}`)
			}},
		{
			name:      "Returns error if url invalid",
			args:      args{"invalid", true, models.Node{}, []byte("{}")},
			portValid: true,
			wantErr:   true,
			want:      nil,
			handleFunc: func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, `{}`)
			}},
		{
			name:      "Returns error if it cannot read response",
			args:      args{"valid", true, models.Node{}, []byte("{}")},
			wantErr:   true,
			portValid: true,
			want:      nil,
			handleFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Length", "1")
			}},
		{
			name:      "Returns error if node returns invalid status code",
			args:      args{"valid", true, models.Node{}, []byte("{}")},
			wantErr:   true,
			want:      nil,
			portValid: true,
			handleFunc: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Error", 404)
			}},
		{
			name:      "Returns error if check batch rpc response returns error",
			args:      args{"valid", true, models.Node{}, []byte(`{}`)},
			wantErr:   true,
			want:      nil,
			portValid: true,
			handleFunc: func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, `{}`)
			}},
		{
			name:      "Returns error if check single rpc response returns error",
			args:      args{"valid", false, models.Node{}, []byte(`{}`)},
			wantErr:   true,
			portValid: true,
			want:      nil,
			handleFunc: func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, `{"error": {"code": -32603}}`)
			}},
		{
			name:      "Returns unmarshaled response if rpc response valid",
			args:      args{"valid", false, models.Node{}, []byte(`{}`)},
			wantErr:   false,
			portValid: true,
			want:      []byte(`{"id": 1}`),
			handleFunc: func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, `{"id": 1}`)
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup()

			if tt.args.url == "valid" && tt.portValid {
				serverURL, _ := url.Parse(server.URL)
				port, _ := strconv.Atoi(serverURL.Port())
				poolerMock.On("GetHTTPPort", mock.Anything).Once().Return(port, nil)
			} else if tt.args.url == "invalid" && tt.portValid {
				poolerMock.On("GetHTTPPort", mock.Anything).Once().Return(1331313, nil)
			} else {
				poolerMock.On("GetHTTPPort", mock.Anything).Once().Return(0, fmt.Errorf("ERROR"))
			}

			mux.HandleFunc("/", tt.handleFunc)

			got, err := SendRequestToNode(tt.args.batch, tt.args.node.ID, tt.args.reqBody)

			if (err != nil) != tt.wantErr {
				t.Errorf("SendRequestToNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if bytes.Compare(got, tt.want) == 1 {
				t.Errorf("SendRequestToNode() = %v, want %v", got, tt.want)
			}

			teardown()
		})
	}
}

func TestCreateRPCError(t *testing.T) {
	type args struct {
		isBatch      bool
		reqRPCBody   RPCRequest
		reqRPCBodies []RPCRequest
		code         int
		message      string
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "Returns single error if it is not batch",
			args: args{false, RPCRequest{ID: 3}, []RPCRequest{}, -32300, "Error"},
			want: RPCResponse{JSONRPC: "2.0", ID: 3, Error: &RPCError{Code: -32300, Message: "Error"}}},
		{
			name: "Returns array of errors if they are batch",
			args: args{true, RPCRequest{}, []RPCRequest{{ID: 3}}, -32300, "Error"},
			want: []RPCResponse{{JSONRPC: "2.0", ID: 3, Error: &RPCError{Code: -32300, Message: "Error"}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateRPCError(tt.args.isBatch, tt.args.reqRPCBody, tt.args.reqRPCBodies, tt.args.code, tt.args.message)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateRPCError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckSingleRPCResponse(t *testing.T) {
	type args struct {
		body []byte
	}
	tests := []struct {
		name    string
		args    args
		want    RPCResponse
		wantErr bool
	}{
		{
			name:    "Returns error if unmarshal fails",
			args:    args{[]byte("INVALID")},
			want:    RPCResponse{},
			wantErr: true},
		{
			name:    "Returns error if rpc code invalid",
			args:    args{[]byte(`{"id": 1, "error": {"code": -32603}}`)},
			want:    RPCResponse{},
			wantErr: true},
		{
			name:    "Returns rpc response if valid",
			args:    args{[]byte(`{"id": 1}`)},
			want:    RPCResponse{ID: 1},
			wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckSingleRPCResponse(tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckSingleRPCResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CheckSingleRPCResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckBatchRPCResponse(t *testing.T) {
	type args struct {
		body []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []RPCResponse
		wantErr bool
	}{
		{
			name:    "Returns error if unmarshal fails",
			args:    args{[]byte("INVALID")},
			want:    nil,
			wantErr: true},
		{
			name:    "Returns rpc response if valid",
			args:    args{[]byte(`[{"id": 1}]`)},
			want:    []RPCResponse{{ID: 1}},
			wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckBatchRPCResponse(tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckBatchRPCResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CheckBatchRPCResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}
