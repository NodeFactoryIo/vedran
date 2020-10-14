package rpc

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/NodeFactoryIo/vedran/internal/models"
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

	type args struct {
		batch   bool
		node    models.Node
		reqBody []byte
	}
	tests := []struct {
		name       string
		args       args
		want       interface{}
		wantErr    bool
		handleFunc handleFnMock
	}{
		{
			name:    "Returns error if node url invalid",
			args:    args{true, models.Node{}, []byte("{}")},
			wantErr: true,
			want:    nil,
			handleFunc: func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, `{}`)
			}},
		{
			name:    "Returns error if it cannot read response",
			args:    args{true, models.Node{}, []byte("{}")},
			wantErr: true,
			want:    nil,
			handleFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Length", "1")
			}},
		{
			name:    "Returns error if node returns invalid status code",
			args:    args{true, models.Node{}, []byte("{}")},
			wantErr: true,
			want:    nil,
			handleFunc: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Error", 404)
			}},
		{
			name:    "Returns error if check batch rpc response returns error",
			args:    args{true, models.Node{}, []byte(`{}`)},
			wantErr: true,
			want:    nil,
			handleFunc: func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, `{}`)
			}},
		{
			name:    "Returns error if check single rpc response returns error",
			args:    args{false, models.Node{}, []byte(`{}`)},
			wantErr: true,
			want:    nil,
			handleFunc: func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, `{"error": {"code": -32603}}`)
			}},
		{
			name:    "Returns unmarshaled response if rpc response valid",
			args:    args{false, models.Node{}, []byte(`{}`)},
			wantErr: false,
			want:    RPCResponse{ID: 1},
			handleFunc: func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, `{"id": 1}`)
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup()

			mux.HandleFunc("/", tt.handleFunc)

			got, err := SendRequestToNode(tt.args.batch, tt.args.node, tt.args.reqBody)

			if (err != nil) != tt.wantErr {
				t.Errorf("SendRequestToNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
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
