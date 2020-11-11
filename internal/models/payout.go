package models

import "time"

type Payout struct {
	ID             string    `storm:"id"`
	Timestamp      time.Time `json:"timestamp"`
	PaymentDetails map[string]NodeStatsDetails
}

type NodeStatsDetails struct {
	TotalPings    float64 `json:"total_pings"`
	TotalRequests float64 `json:"total_requests"`
}
