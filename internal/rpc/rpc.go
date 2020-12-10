package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/NodeFactoryIo/vedran/internal/configuration"
)

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RPCResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      uint64           `json:"id"`
	Result  *json.RawMessage `json:"result,omitempty"`
	Error   *RPCError        `json:"error,omitempty"`
}

type RPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      uint64      `json:"id"`
	Method  string      `json:"method,omitempty"`
	Params  interface{} `json:"params,omitempty"`
}

const (
	InternalServerError = -32603
	ParseError          = -32700
	InvalidRequest      = -32600

	RequestTimeout = 3 * time.Second
)

// IsBatch returns if request contains batch rpc requests
func IsBatch(reqBody []byte) bool {
	x := bytes.TrimLeft(reqBody, " \t\r\n")
	if len(x) > 0 && x[0] == '[' {
		return true
	}

	return false
}

func createSingleRPCError(id uint64, code int, message string) RPCResponse {
	return RPCResponse{
		ID: id,
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
		JSONRPC: "2.0",
	}
}

// CreateRPCError returns rpc errors for appropriate request ids
func CreateRPCError(isBatch bool, reqRPCBody RPCRequest, reqRPCBodies []RPCRequest, code int, message string) interface{} {
	if !isBatch {
		return createSingleRPCError(reqRPCBody.ID, code, message)
	}

	rpcResponses := make([]RPCResponse, len(reqRPCBodies))
	for i, body := range reqRPCBodies {
		rpcResponses[i] = createSingleRPCError(body.ID, code, message)
	}
	return rpcResponses
}

// CheckSingleRPCResponse checks for errors in non batch rpc response
func CheckSingleRPCResponse(body []byte) (RPCResponse, error) {
	var rpcResponse RPCResponse

	err := json.Unmarshal(body, &rpcResponse)
	if err != nil {
		return RPCResponse{}, err
	} else if rpcResponse.Error != nil {
		if rpcResponse.Error.Code == InternalServerError {
			return RPCResponse{}, fmt.Errorf("Invalid rpc code %d", InternalServerError)
		}
	}

	return rpcResponse, nil
}

// CheckBatchRPCResponse checks for errors in batch rpc response
func CheckBatchRPCResponse(body []byte) ([]RPCResponse, error) {
	var rpcResponses []RPCResponse

	err := json.Unmarshal(body, &rpcResponses)
	if err != nil {
		return nil, err
	}

	return rpcResponses, nil
}

// SendRequestToNode routes request to node and checks response
func SendRequestToNode(isBatch bool, nodeID string, reqBody []byte) ([]byte, error) {
	port, err := configuration.Config.PortPool.GetHTTPPort(nodeID)
	if err != nil {
		return nil, err
	}

	client := http.Client{
		Timeout: RequestTimeout,
	}
	resp, err := client.Post(
		"http://127.0.0.1:"+strconv.Itoa(port)+"/",
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status code is not 200")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if isBatch {
		_, err = CheckBatchRPCResponse(body)
	} else {
		_, err = CheckSingleRPCResponse(body)
	}

	if err != nil {
		return nil, err
	}

	return body, nil
}
