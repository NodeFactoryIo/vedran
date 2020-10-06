package rpc

import (
	"encoding/json"
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
)

// IsBatch returns if requst contains batch rpc requests
func IsBatch(reqBody RPCRequest) bool {
	if reqBody != (RPCRequest{}) {
		return false
	}

	return true
}

// Unmarshal stores json data in appropriate interface reqRPCBody if it is not batch
// and reqRPCBodies if request is batched
func Unmarshal(body []byte, reqRPCBody interface{}, reqRPCBodies interface{}) error {
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
func CreateRPCError(reqRPCBody RPCRequest, reqRPCBodies []RPCRequest, message string) interface{} {
	if !IsBatch(reqRPCBody) {
		return RPCResponse{
			ID: reqRPCBody.ID,
			Error: &RPCError{
				Code:    InternalServerError,
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
				Code:    InternalServerError,
				Message: message,
			},
			JSONRPC: "2.0"}
		i++
	}
	return rpcResponses
}
