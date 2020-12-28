package controllers

import (
	"net/http"

	"github.com/NodeFactoryIo/vedran/internal/configuration"
	"github.com/NodeFactoryIo/vedran/internal/ws"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (c ApiController) WSHandler(w http.ResponseWriter, r *http.Request) {
	nodes := c.repositories.NodeRepo.GetActiveNodes(configuration.Config.Selection)
	if len(*nodes) == 0 {
		log.Error("Request failed because vedran has no available nodes")
		http.Error(w, "No available nodes", 503)
		return
	}

	connToLoadbalancer, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Failed upgrading connection because of %v", err)
		http.Error(w, "Failed upgrading connection", 500)
		return
	}

	connErr := make(chan *ws.ConnectionError)
	messages := make(chan ws.Message)
	wsConnection := make(chan *websocket.Conn)
	for _, node := range *nodes {
		go ws.EstablishNodeConn(node.ID, wsConnection, messages, connErr)

		connectionError := <-connErr
		connToNode := <-wsConnection
		if connectionError != nil {
			log.Errorf("Establishing connection failed because of %v", err)
			if connectionError.IsNodeError() {
				c.actions.PenalizeNode(node, c.repositories)
			}
			continue
		}

		go c.repositories.NodeRepo.UpdateNodeUsed(node)

		go ws.SendRequestToNode(connToLoadbalancer, connToNode, node, c.repositories, c.actions)
		go ws.SendResponseToClient(connToLoadbalancer, connToNode, messages, node, c.repositories)
		return
	}

	log.Error("Failed establishing connection with any node")
	_ = connToLoadbalancer.Close()
	close(connErr)
	close(messages)
	close(wsConnection)
}
