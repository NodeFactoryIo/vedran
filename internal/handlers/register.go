package handlers

import (
	"errors"
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/db"
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

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var p RegisterRequest

	err := util.DecodeJSONBody(w, r, &p)
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

	database := db.GetDatabaseService()
	node := db.Node{
		ID:            p.Id,
		ConfigHash:    p.ConfigHash,
		NodeUrl:       p.NodeUrl,
		PayoutAddress: p.PayoutAddress,
		Token:         ,
	}
	database.DB.Save()

	_, _ = fmt.Fprintf(w, "Register request: %+v", p)
}
