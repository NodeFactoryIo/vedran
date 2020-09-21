package controllers

import (
	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"log"
	"net/http"
)

func (c ApiController) PingHandler(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(auth.RequestContextKey).(*auth.RequestContext)
	err := c.pingRepo.Save(&models.Ping{
		NodeId:    request.NodeId,
		Timestamp: request.Timestamp,
	})
	if err != nil {
		// error on saving in database
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}