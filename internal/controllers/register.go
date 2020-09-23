package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/pkg/util"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type RegisterRequest struct {
	Id            string `json:"id"`
	ConfigHash    string `json:"config_hash"`
	NodeUrl       string `json:"node_url"`
	PayoutAddress string `json:"payout_address"`
}

type RegisterResponse struct {
	Token string `json:"token"`
}

func (c ApiController) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// decode request body
	var registerRequest RegisterRequest
	err := util.DecodeJSONBody(w, r, &registerRequest)
	if err != nil {
		var mr *util.MalformedRequest
		if errors.As(err, &mr) {
			log.Errorf("Malformed request error: %v", err)
			http.Error(w, mr.Msg, mr.Status)
		} else {
			// unknown error
			log.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	if c.whitelistEnabled {
		_, err := c.nodeRepo.IsNodeWhitelisted(registerRequest.Id)
		if err != nil {
			log.Errorf("Node id %s not whitelisted: %v", registerRequest.Id, err)
			http.Error(w, fmt.Sprintf("Node %s is not whitelisted", registerRequest.Id), http.StatusBadRequest)
			return
		}
	}

	// generate auth token
	token, err := auth.CreateNewToken(registerRequest.Id)
	if err != nil {
		// unknown error
		log.Errorf("Unable to create auth token, error: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// save node to database
	node := &models.Node{
		ID:            registerRequest.Id,
		ConfigHash:    registerRequest.ConfigHash,
		NodeUrl:       registerRequest.NodeUrl,
		PayoutAddress: registerRequest.PayoutAddress,
		Token:         token,
	}
	err = c.nodeRepo.Save(node)
	if err != nil {
		log.Errorf("Unable to save node %v to database, error: %v", node, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	log.Infof("New node %s registered", node.ID)

	// return generated token
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(RegisterResponse{token})
}
