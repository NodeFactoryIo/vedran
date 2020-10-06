package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/NodeFactoryIo/vedran/internal/models"
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
)

// IsBatch returns if requst contains batch rpc requests
func IsBatch(reqBody RPCRequest) bool {
	if reqBody != (RPCRequest{}) {
		return false
	}

	return true
}

// UnmarshalRequest stores json data in appropriate interface reqRPCBody if it is not batch
// and reqRPCBodies if request is batched
func UnmarshalRequest(body []byte, reqRPCBody *RPCRequest, reqRPCBodies *[]RPCRequest) error {
	err := json.Unmarshal(body, &reqRPCBody)
	if err != nil {
		err = json.Unmarshal(body, &reqRPCBodies)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateRPCError returns rpc errors for appropriate request ids
func CreateRPCError(reqRPCBody RPCRequest, reqRPCBodies []RPCRequest, code int, message string) interface{} {
	if !IsBatch(reqRPCBody) {
		return RPCResponse{
			ID: reqRPCBody.ID,
			Error: &RPCError{
				Code:    code,
				Message: message,
			},
			JSONRPC: "2.0",
		}
	}

	rpcResponses := make([]RPCResponse, len(reqRPCBodies))
	for i, body := range reqRPCBodies {
		rpcResponses[i] = RPCResponse{
			ID: body.ID,
			Error: &RPCError{
				Code:    code,
				Message: message,
			},
			JSONRPC: "2.0"}
		i++
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

// SendRequestToNode routes request to node and checks responss
func SendRequestToNode(node models.Node, reqBody []byte, reqRPCBody RPCRequest) (interface{}, error) {
	resp, err := http.Post(node.NodeUrl, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Status code is not 200")
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var rpcResponse interface{}
	if IsBatch(reqRPCBody) {
		rpcResponse, err = CheckBatchRPCResponse(body)
	} else {
		rpcResponse, err = CheckSingleRPCResponse(body)
	}

	if err != nil {
		return nil, err
	}

	return rpcResponse, nil
}
