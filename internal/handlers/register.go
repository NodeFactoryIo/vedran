package handlers

import (
	"encoding/json"
	"errors"
	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/internal/db"
	"github.com/NodeFactoryIo/vedran/pkg/util"
	"log"
	"net/http"
	"strconv"
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

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
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

	// generate auth token
	token, err := auth.CreateNewToken(registerRequest.Id)
	if err != nil {
		// unknown error
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	database := db.GetDatabaseService()
	id, err := strconv.Atoi(registerRequest.Id)
	node := db.Node{
		ID:            id,
		ConfigHash:    registerRequest.ConfigHash,
		NodeUrl:       registerRequest.NodeUrl,
		PayoutAddress: registerRequest.PayoutAddress,
		Token:         token,
	}
	err = database.DB.Save(node)

	// return generated token
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(RegisterResponse{token})
}
