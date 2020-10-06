package controllers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/NodeFactoryIo/vedran/internal/rpc"
	log "github.com/sirupsen/logrus"
)

func (c ApiController) RPCHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	defer r.Body.Close()
	reqBody, _ := ioutil.ReadAll(r.Body)
	var reqRPCBody rpc.RPCRequest
	var reqRPCBodies []rpc.RPCRequest
	err := rpc.Unmarshal(reqBody, &reqRPCBody, &reqRPCBodies)
	if err != nil {
		log.Error("Request failed because of: %v", err)
		_ = json.NewEncoder(w).Encode(
			rpc.CreateRPCError(reqRPCBody, reqRPCBodies, err.Error()))
		return
	}

	nodes, _ := c.nodeRepo.GetActiveNodes()
	if len(*nodes) == 0 {
		log.Error("Request failed because vedran has no available nodes")
		_ = json.NewEncoder(w).Encode(
			rpc.CreateRPCError(reqRPCBody, reqRPCBodies, "No available nodes"))
		return
	}

	// @TODO: Peer selection code

	for _, node := range *nodes {
		resp, err := http.Post(node.NodeUrl, "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			log.Errorf("Request failed to %s because of: %v", node.ID, err)
			continue
		} else if resp.StatusCode != 200 {
			log.Error("Request failed because status code was not 200")
			continue
		}

		defer resp.Body.Close()
		isBatch := rpc.IsBatch(reqRPCBody)
		body, _ := ioutil.ReadAll(resp.Body)
		if isBatch {
			var rpcResponses []rpc.RPCResponse

			err = json.Unmarshal(body, &rpcResponses)
			if err != nil {
				log.Errorf("Request failed to %s because of: %v", node.ID, err)
				continue
			}

			_ = json.NewEncoder(w).Encode(rpcResponses)
		} else {
			var rpcResponse rpc.RPCResponse

			err = json.Unmarshal(body, &rpcResponse)
			if err != nil {
				log.Errorf("Request failed to %s because of: %v", node.ID, err)
				continue
			} else if rpcResponse.Error != nil {
				if rpcResponse.Error.Code == rpc.InternalServerError {
					log.Errorf("Request failed to %s because of invalid code: %d", node.ID, rpc.InternalServerError)
					continue
				}
			}

			_ = json.NewEncoder(w).Encode(rpcResponse)
		}

		return
	}

	log.Error("Request failed because all nodes returned invalid rpc response")
	_ = json.NewEncoder(w).Encode(
		rpc.CreateRPCError(reqRPCBody, reqRPCBodies, "Internal Server Error"))
}
