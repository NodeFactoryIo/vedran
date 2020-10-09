package models

import "time"

type Record struct {
	NodeId    string `storm:"id"`
	Status    string
	Timestamp time.Time
}

type RecordRepository interface {
	Save(record *Record) error
}
