package ws

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/NodeFactoryIo/vedran/internal/configuration"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Message struct {
	msg     []byte
	msgType int
}

var (
	origin = "http://localhost/"
)

// SendRequestToNode reads incoming messages to load balancer and pipes
// them to node
func SendRequestToNode(conn *websocket.Conn, nodeConn *websocket.Conn) {
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		nodeConn.WriteMessage(msgType, msg)
	}
}

// SendResponseToClient iterates through messages sent from node connection and sends them
// to client
func SendResponseToClient(conn *websocket.Conn, nodeConn *websocket.Conn, messages chan Message) {
	for m := range messages {
		if err := conn.WriteMessage(m.msgType, m.msg); err != nil {
			log.Errorf("Establishing node connection failed because of %v:", err)
			return
		}
	}
}

// EstablishNodeConn dials node, pipes messages to message channel and returns connection
// to wsConnection channel
func EstablishNodeConn(nodeID string, wsConnection chan *websocket.Conn, messages chan Message, connErr chan error) {
	port, err := configuration.Config.PortPool.GetWSPort(nodeID)
	if err != nil {
		connErr <- err
		return
	}

	host, _ := url.Parse("ws://127.0.0.1:" + strconv.Itoa(port))
	c, _, err := websocket.DefaultDialer.Dial(host.String(), http.Header{"Origin": {origin}})
	if err != nil {
		connErr <- err
		return
	}

	wsConnection <- c

	defer c.Close()
	for {
		msgType, m, err := c.ReadMessage()
		if err != nil {
			connErr <- err
			log.Errorf("Failed reading message from node because of %v:", err)
			return
		}

		messages <- Message{msgType: msgType, msg: m}
	}
}
