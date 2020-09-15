package handlers

import (
	"encoding/json"
	"errors"
	"github.com/NodeFactoryIo/vedran/internal/auth"
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

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// decode body
	var registerRequest RegisterRequest
	err := util.DecodeJSONBody(w, r, &registerRequest)
	if err != nil {
		var mr *util.MalformedRequest
		if errors.As(err, &mr) {
			// malformed request err2
			http.Error(w, mr.Msg, mr.Status)
		} else {
			// unknown err2
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	token, err := auth.CreateNewToken()
	if err != nil {
		// unknown err2
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(RegisterResponse{
		token,
	})
}
