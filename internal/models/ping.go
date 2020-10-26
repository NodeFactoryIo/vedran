package models

import "time"

type Ping struct {
	NodeId    string `storm:"id"`
	Timestamp time.Time
}
