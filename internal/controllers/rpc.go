package controllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/NodeFactoryIo/vedran/internal/configuration"
	"github.com/NodeFactoryIo/vedran/internal/record"
	"github.com/NodeFactoryIo/vedran/internal/rpc"
	log "github.com/sirupsen/logrus"
)

func (c ApiController) RPCHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	defer r.Body.Close()
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Request failed because of: %v", err)
		_ = json.NewEncoder(w).Encode(
			rpc.CreateRPCError(false, rpc.RPCRequest{}, nil, rpc.ParseError, "Parse error"))
		return
	}

	var reqRPCBody rpc.RPCRequest
	var reqRPCBodies []rpc.RPCRequest
	isBatch := rpc.IsBatch(reqBody)
	if isBatch {
		err = json.Unmarshal(reqBody, &reqRPCBodies)
	} else {
		err = json.Unmarshal(reqBody, &reqRPCBody)
	}

	if err != nil {
		log.Errorf("Request failed because of: %v", err)
		_ = json.NewEncoder(w).Encode(
			rpc.CreateRPCError(isBatch, reqRPCBody, reqRPCBodies, rpc.ParseError, "Parse error"))
		return
	}

	nodes := c.repositories.NodeRepo.GetActiveNodes(configuration.Config.Selection)
	if len(*nodes) == 0 {
		log.Error("Request failed because vedran has no available nodes")
		_ = json.NewEncoder(w).Encode(
			rpc.CreateRPCError(isBatch, reqRPCBody, reqRPCBodies, rpc.InternalServerError, "No available nodes"))
		return
	}

	for _, node := range *nodes {
		byteResponse, err := rpc.SendRequestToNode(
			isBatch,
			node.ID,
			reqBody,
		)
		if err != nil {
			log.Errorf("Request failed to node %s because of: %v", node.ID, err)
			go record.FailedRequest(node, c.repositories, c.actions)
			continue
		}

		go record.SuccessfulRequest(node, c.repositories, c.actions)
		_, _ = w.Write(byteResponse)
		return
	}

	log.Error("Request failed because all nodes returned invalid rpc response")
	_ = json.NewEncoder(w).Encode(
		rpc.CreateRPCError(isBatch, reqRPCBody, reqRPCBodies, rpc.InternalServerError, "Internal Server Error"))
}
