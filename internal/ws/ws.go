package ws

import (
	"github.com/NodeFactoryIo/vedran/internal/actions"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/record"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"net/url"
	"strconv"
	"time"

	"github.com/NodeFactoryIo/vedran/internal/configuration"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Message struct {
	msg     []byte
	msgType int
}

var (
	ShortHandshakeTimeout = 2 * time.Second
)

// SendRequestToNode reads incoming messages to load balancer and pipes
// them to node
func SendRequestToNode(conn *websocket.Conn, nodeConn *websocket.Conn) {
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Errorf("Reading request from client failed because of %v:", err)
			return
		}
		_ = nodeConn.WriteMessage(msgType, msg)
	}
}

// SendResponseToClient iterates through messages sent from node connection and sends them
// to client
func SendResponseToClient(conn *websocket.Conn, nodeConn *websocket.Conn, messages chan Message, a actions.Actions, repositories repositories.Repos, node models.Node) {
	for m := range messages {
		if err := conn.WriteMessage(m.msgType, m.msg); err != nil {
			log.Errorf("Sending response client failed because of %v:", err)
			return
		}
		record.SuccessfulRequest(node, repositories, a)
	}
}

// EstablishNodeConn dials node, pipes messages to message channel and returns connection
// to wsConnection channel
func EstablishNodeConn(nodeID string, wsConnection chan *websocket.Conn, messages chan Message, connErr chan *ConnectionError) {
	port, err := configuration.Config.PortPool.GetWSPort(nodeID)
	if err != nil {
		connErr <- &ConnectionError{
			Err:  err,
			Type: PortPoolError,
		}
		wsConnection <- nil
		return
	}

	host, _ := url.Parse("ws://127.0.0.1:" + strconv.Itoa(port))
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = ShortHandshakeTimeout
	c, _, err := dialer.Dial(host.String(), nil) // http.Header{"Origin": {origin}}
	if err != nil {
		connErr <- &ConnectionError{
			Err:  err,
			Type: NodeError,
		}
		wsConnection <- nil
		return
	}

	connErr <- nil
	wsConnection <- c

	defer c.Close()
	for {
		msgType, m, err := c.ReadMessage()
		if err != nil {
			log.Errorf("Failed reading message from node because of %v:", err)
			return
		}

		messages <- Message{msgType: msgType, msg: m}
	}
}
