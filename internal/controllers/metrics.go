package controllers

import (
	"errors"
	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/pkg/util"
	"log"
	"net/http"
)

type MetricsRequest struct {
	PeerCount             int32 `json:"peer_count"`
	BestBlockHeight       int64 `json:"best_block_height"`
	FinalizedBlockHeight  int64 `json:"finalized_block_height"`
	ReadyTransactionCount int32 `json:"ready_transaction_count"`
}

func (c ApiController) SaveMetricsHandler(w http.ResponseWriter, r *http.Request) {
	// decode request body
	var metricsRequest MetricsRequest
	err := util.DecodeJSONBody(w, r, &metricsRequest)
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

	requestContext := r.Context().Value(auth.RequestContextKey).(*auth.RequestContext)

	err = c.metricsRepo.Save(&models.Metrics{
		NodeId:                requestContext.NodeId,
		PeerCount:             metricsRequest.PeerCount,
		BestBlockHeight:       metricsRequest.BestBlockHeight,
		FinalizedBlockHeight:  metricsRequest.FinalizedBlockHeight,
		ReadyTransactionCount: metricsRequest.ReadyTransactionCount,
	})

	if err != nil {
		// error on saving in database
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
