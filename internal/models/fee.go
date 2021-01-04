package models

type Fee struct {
	NodeId   string `storm:"id"`
	TotalFee int64  `json:"total_fee"`
}
