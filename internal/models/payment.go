package models

import "time"

type Payment struct {
	Timestamp      time.Time `json:"timestamp"`
	PaymentDetails map[string]NodePaymentDetails
}

type NodePaymentDetails struct {
	TotalPings    float64 `json:"total_pings"`
	TotalRequests float64 `json:"total_requests"`
}
