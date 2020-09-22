package controllers

import (
	"errors"
	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/pkg/util"
	log "github.com/sirupsen/logrus"
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
			log.Errorf("Malformed request error: %v", err)
			http.Error(w, mr.Msg, mr.Status)
		} else {
			// unknown error
			log.Error(err)
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
		log.Errorf("Unable to save metrics for node %v to database, error: %v", requestContext.NodeId, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	log.Debugf("Node %s saved new metrics", requestContext.NodeId)
}
