package models

type Metrics struct {
	NodeId                string `storm:"id"`
	PeerCount             int32
	BestBlockHeight       int64
	FinalizedBlockHeight  int64
	ReadyTransactionCount int32
}

type MetricsRepository interface {
	FindByID(ID string) (*Metrics, error)
	Save(metrics *Metrics) error
	GetAll() (*[]Metrics, error)
}
