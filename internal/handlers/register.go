package handlers

import (
	"errors"
	"fmt"
	"github.com/NodeFactoryIo/vedran/pkg/util"
	"github.com/dgrijalva/jwt-go"
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
			// malformed request error
			http.Error(w, mr.Msg, mr.Status)
		} else {
			// unknown error
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	// create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"authorized": true,
		"user_id": "1",
	})
	// TODO -> move secret to env variable
	stringToken, error := token.SignedString([]byte("jdnfksdmfksd"))
	if error != nil {
		// unknown error
		log.Println(error.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	t := RegisterResponse{
		Token: stringToken,
	}

	_, _ = fmt.Fprintf(w, "%+v", t)
}
