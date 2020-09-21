package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/pkg/util"
	"log"
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
			// malformed request error
			http.Error(w, mr.Msg, mr.Status)
		} else {
			// unknown error
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	if c.whitelistEnabled {
		whitelisted, err := c.nodeRepo.IsNodeWhitelisted(registerRequest.Id)
		if err != nil {
			// error on querying database
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if !whitelisted {
			log.Printf("Node id %s not whitelisted", registerRequest.Id)
			http.Error(w, fmt.Sprintf("Node %s is not whitelisted", registerRequest.Id), http.StatusBadRequest)
			return
		}
	}

	// generate auth token
	token, err := auth.CreateNewToken(registerRequest.Id)
	if err != nil {
		// unknown error
		log.Println(err.Error())
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
		// error on saving in database
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// return generated token
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(RegisterResponse{token})
}