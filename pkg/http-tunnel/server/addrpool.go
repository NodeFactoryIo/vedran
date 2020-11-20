package server

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type RemoteID struct {
	ClientID string
	PortName string
	Port     int
}

type AddrPool struct {
	first   int
	last    int
	used    int
	mutex   sync.Mutex
	addrMap map[int]*RemoteID
}

type Pooler interface {
	Init(rang string) error
	Acquire(cname string, pname string) (int, error)
	Release(id string) error
	GetHTTPPort(id string) (int, error)
	GetWSPort(id string) (int, error)
}

func (ap *AddrPool) Init(rang string) error {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	rarray := strings.Split(rang, ":")
	if len(rarray) != 2 {
		return fmt.Errorf("Port Range Bad Formated %s", rang)
	}

	ap.first, _ = strconv.Atoi(rarray[0])
	ap.last, _ = strconv.Atoi(rarray[1])

	if ap.last < ap.first {
		return fmt.Errorf("Port Range Bad Formated  last %d slower than first %d", ap.last, ap.first)
	}

	ap.addrMap = make(map[int]*RemoteID)

	return nil
}

func (ap *AddrPool) Acquire(cname string, pname string) (int, error) {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()
	assignedPort := 0
	// search for the first unnused port
	for i := ap.first; i < ap.last; i++ {
		cur := ap.addrMap[i]
		if cur == nil {
			//empty
			assignedPort = i
			ap.used++
			ap.addrMap[i] = &RemoteID{
				ClientID: cname,
				PortName: pname,
				Port:     assignedPort,
			}
			break
		}
	}
	if assignedPort == 0 {
		return 0, fmt.Errorf("pool is full , can not assign any Port Address")
	}
	return assignedPort, nil
}

func (ap *AddrPool) Release(id string) error {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()
	found := false
	// search for the first unnused port
	for i := ap.first; i < ap.last; i++ {
		cur := ap.addrMap[i]
		if cur != nil && cur.ClientID == id {
			//empty
			ap.used--
			ap.addrMap[i] = nil
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("ID %s not found in pool", id)
	}

	return nil
}

// GetHTTPPort retrieves port for given node id http tunnel
func (ap *AddrPool) GetHTTPPort(id string) (int, error) {
	for _, addr := range ap.addrMap {
		if addr != nil && addr.ClientID == id && addr.PortName == "http" {
			return addr.Port, nil
		}
	}

	return 0, fmt.Errorf("No port for id %s in pool", id)
}

// GetWSPort retrieves port for given node id websocket tunel
func (ap *AddrPool) GetWSPort(id string) (int, error) {
	for _, addr := range ap.addrMap {
		if addr != nil && addr.ClientID == id && addr.PortName == "ws" {
			return addr.Port, nil
		}
	}

	return 0, fmt.Errorf("No port for id %s in pool", id)
}
