package ws

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/NodeFactoryIo/vedran/internal/actions"
	"github.com/NodeFactoryIo/vedran/internal/configuration"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/record"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
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
func SendRequestToNode(
	connToLoadbalancer *websocket.Conn,
	connToNode *websocket.Conn,
	node models.Node,
	repos repositories.Repos,
	act actions.Actions,
) {
	for {
		msgType, msg, err := connToLoadbalancer.ReadMessage()
		if err != nil {
			log.Errorf("Reading request from client failed because of %v:", err)
			closeConnections(connToLoadbalancer, connToNode, node)
			return
		}
		err = connToNode.WriteMessage(msgType, msg)
		if err != nil {
			record.FailedRequest(node, repos, act)
			closeConnections(connToLoadbalancer, connToNode, node)
			return
		}
	}
}

// SendResponseToClient iterates through messages sent from node connection and sends them
// to client
func SendResponseToClient(
	connToLoadbalancer *websocket.Conn,
	connToNode *websocket.Conn,
	messages chan Message,
	node models.Node,
	repos repositories.Repos,
) {
	for m := range messages {
		if err := connToLoadbalancer.WriteMessage(m.msgType, m.msg); err != nil {
			log.Errorf("Sending response client failed because of %v:", err)
			closeConnections(connToLoadbalancer, connToNode, node)
			return
		}
		record.SuccessfulRequest(node, repos)
	}
}

func closeConnections(connToLoadbalancer *websocket.Conn, connToNode *websocket.Conn, node models.Node) {
	closeConn(connToLoadbalancer, "error on closing ws connection towards loadbalancer")
	closeConn(connToNode, fmt.Sprintf("error on closing ws connection towards node %s", node.ID))
}

func closeConn(conn *websocket.Conn, errorMessage string) {
	err := conn.Close()
	if err != nil {
		log.Errorf(errorMessage)
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
	c, _, err := dialer.Dial(host.String(), nil)
	if err != nil {
		if strings.Contains(err.Error(), "cancel") {
			connErr <- &ConnectionError{
				Err:  err,
				Type: UserCancellationError,
			}
		} else {
			connErr <- &ConnectionError{
				Err:  err,
				Type: NodeError,
			}
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
			closeConn(c, fmt.Sprintf("error on closing ws connection towards node %s", nodeID))
			return
		}

		messages <- Message{msgType: msgType, msg: m}
	}
}
