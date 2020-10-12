// Copyright (C) 2017 Michał Matczuk
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package server

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"sync"
)

// RegistryItem holds information about hosts and listeners associated with a
// client.
type RegistryItem struct {
	Hosts         []*HostAuth
	Listeners     []net.Listener
	ListenerNames []string
	ClientName    string
	ClientID      string
}

// HostAuth holds host and authentication info.
type HostAuth struct {
	Host string
	Auth *Auth
}

type hostInfo struct {
	identifier string
	auth       *Auth
}

type registry struct {
	source map[string]*RegistryItem //Origin Address based on host:port
	items  map[string]*RegistryItem //Client name
	hosts  map[string]*hostInfo
	mu     sync.RWMutex
	logger *log.Entry
}

func newRegistry(logger *log.Entry) *registry {
	if logger == nil {
		logger = log.NewEntry(log.StandardLogger())
	}

	return &registry{
		items:  make(map[string]*RegistryItem),
		source: make(map[string]*RegistryItem),
		hosts:  make(map[string]*hostInfo),
		logger: logger,
	}
}

func (r *registry) PreSubscribe(origaddr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.source[origaddr]; ok {
		r.logger.Errorf("error on pre-subscribe to registry this entry already exist %s", origaddr)
		return
	}
	r.source[origaddr] = &RegistryItem{ClientID: origaddr}
}

// Subscribe allows to connect client with a given identifier.
func (r *registry) Subscribe(cname string, origaddr string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[cname]; ok {
		r.logger.Errorf("error on subscribe to registry this entry already exist %s", origaddr)
		return
	}
	reg := r.source[origaddr]
	reg.ClientName = cname
	r.items[cname] = reg
	r.logger.WithFields(log.Fields{
		"client-name": cname,
		"client-id":   origaddr,
		"data":        reg,
	}).Info("REGISTRY SUBSCRIBE")
}

// GetID returns the ID for this client
func (r *registry) GetID(cname string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.items[cname]
	if !ok {
		return ""
	}
	return v.ClientID
}

// Subscriber returns client identifier assigned to given host.
func (r *registry) Subscriber(hostPort string) (string, *Auth, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.hosts[trimPort(hostPort)]
	if !ok {
		return "", nil, false
	}
	fmt.Printf("SUBSCRIBER REGISTRY [%s] value : %+v \n", hostPort, h)

	return h.identifier, h.auth, ok
}

// Unsubscribe removes client from registry and returns it's RegistryItem.
func (r *registry) Unsubscribe(identifier string, idname string) *RegistryItem {
	r.mu.Lock()
	defer r.mu.Unlock()

	i, ok := r.items[identifier]
	if !ok {
		fmt.Printf("UNSUBSCRIBE REGISTRY error not found ID [%s] Idname [%s] value : %+v \n", identifier, idname, i)
		return nil
	}
	fmt.Printf("UNSUBSCRIBE REGISTRY Identifier [%s] Idname [%s] value : %+v \n", identifier, idname, i)

	r.logger.WithFields(log.Fields{
		"identifier": identifier,
		"id-name":    idname,
		"data":       i,
	}).Info("REGISTRY UNSUBSCRIBE")

	if i.Hosts != nil {
		for _, h := range i.Hosts {
			delete(r.hosts, h.Host)
		}
	}

	delete(r.items, identifier)

	return i
}

func (r *registry) set(i *RegistryItem, identifier string) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	j, ok := r.items[identifier]
	if !ok {
		r.logger.WithFields(log.Fields{
			"client-name": identifier,
			"client-id":   i.ClientID,
			"data":        i,
		}).Error("REGISTRY SET ERROR: client-name not found")

		return errClientNotSubscribed
	}

	r.logger.WithFields(log.Fields{
		"client-name": identifier,
		"client-id":   j.ClientID,
		"data":        j,
	}).Info("REGISTRY SET (OLD FOUND)")

	i.ClientID = j.ClientID

	r.logger.WithFields(log.Fields{
		"client-name": identifier,
		"client-id":   i.ClientID,
		"data":        i,
	}).Info("REGISTRY SET (NEW SET)")

	if i.Hosts != nil {
		for _, h := range i.Hosts {
			if h.Auth != nil && h.Auth.Token == "" {
				return fmt.Errorf("missing auth token")
			}
			if _, ok := r.hosts[trimPort(h.Host)]; ok {
				return fmt.Errorf("host %q is occupied", h.Host)
			}
		}

		for _, h := range i.Hosts {
			r.hosts[trimPort(h.Host)] = &hostInfo{
				identifier: identifier,
				auth:       h.Auth,
			}
		}
	}
	r.items[identifier] = i
	r.source[i.ClientID] = i

	return nil
}

func (r *registry) clear(identifier string) *RegistryItem {

	r.mu.Lock()
	defer r.mu.Unlock()

	ilogger := r.logger.WithFields(log.Fields{
		"identifier": identifier,
	})

	i, ok := r.source[identifier]
	if !ok || i == nil {
		ilogger.WithFields(log.Fields{
			"register-exist": ok,
		}).Error("error on clear register")
		return nil
	}

	iccloger := ilogger.WithFields(log.Fields{
		"client-name": i.ClientName,
		"client-id":   i.ClientID,
	})

	iccloger.WithFields(log.Fields{
		"data": i,
	}).Debug("REGISTRY CLEAR item")

	if i.Hosts != nil {
		for _, h := range i.Hosts {
			iccloger.WithFields(log.Fields{
				"host":     h.Host,
				"trimport": trimPort(h.Host),
			}).Debug("REGISTRI CLEAR (delete hosts)")
			delete(r.hosts, trimPort(h.Host))
		}
	}

	r.source[identifier] = nil
	r.items[i.ClientName] = nil
	delete(r.source, identifier)
	delete(r.items, i.ClientName)
	return i
}

func trimPort(hostPort string) (host string) {
	host, _, _ = net.SplitHostPort(hostPort)
	if host == "" {
		host = hostPort
	}
	return
}
