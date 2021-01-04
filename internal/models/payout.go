package models

import "time"

type Payout struct {
	ID             int       `storm:"id,increment"`
	Timestamp      time.Time `json:"timestamp"`
	PaymentDetails map[string]NodeStatsDetails
	LbFee          float64 `json:"lb_fee"`
}

type NodeStatsDetails struct {
	TotalPings    float64 `json:"total_pings"`
	TotalRequests float64 `json:"total_requests"`
}
