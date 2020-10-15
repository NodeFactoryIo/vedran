package controllers

import (
	"encoding/json"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/NodeFactoryIo/vedran/internal/configuration"
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
		rpcResponse, err := rpc.SendRequestToNode(isBatch, node, reqBody)
		if err != nil {
			log.Errorf("Request failed to node %s because of: %v", node.ID, err)
			// start penalize node action
			go c.actions.PenalizeNode(node, c.repositories)
			// save failed record
			err := c.repositories.RecordRepo.Save(&models.Record{
				NodeId:    node.ID,
				Timestamp: time.Now(),
				Status:    "failed",
			})
			if err != nil {
				log.Errorf("Failed saving failed request because of: %v", err)
			}
			continue
		}

		// start reward node action
		go c.actions.RewardNode(node, c.repositories)
		// save successful record
		err = c.repositories.RecordRepo.Save(&models.Record{
			NodeId:    node.ID,
			Timestamp: time.Now(),
			Status:    "successful",
		})
		if err != nil {
			log.Errorf("Failed saving successful request because of: %v", err)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse)
		return
	}

	log.Error("Request failed because all nodes returned invalid rpc response")
	_ = json.NewEncoder(w).Encode(
		rpc.CreateRPCError(isBatch, reqRPCBody, reqRPCBodies, rpc.InternalServerError, "Internal Server Error"))
}
