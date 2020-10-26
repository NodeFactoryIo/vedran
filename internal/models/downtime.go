package models

import "time"

type Downtime struct {
	ID     int `storm:"id,increment"`
	NodeId string
	End    time.Time
	Start  time.Time
}
