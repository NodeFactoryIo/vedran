package models

import "time"

type Ping struct {
	NodeId    string `storm:"id"`
	Timestamp time.Time
}

type PingRepository interface {
	FindByNodeID(nodeId string) (*Ping, error)
	Save(ping *Ping) error
	GetAll() (*[]Ping, error)
	// Calculates last ping time and downtime duration
	CalculateDowntime(nodeId string, pingTime time.Time) (time.Time, time.Duration, error)
}
