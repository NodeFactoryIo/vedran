package repositories

// Repos structure holds all available repositories
type Repos struct {
	NodeRepo     NodeRepository
	PingRepo     PingRepository
	MetricsRepo  MetricsRepository
	RecordRepo   RecordRepository
	DowntimeRepo DowntimeRepository
	PaymentRepo  PaymentRepository
}
