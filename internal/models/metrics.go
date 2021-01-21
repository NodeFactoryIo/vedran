package models

type Metrics struct {
	NodeId                string `storm:"id"`
	PeerCount             int32
	BestBlockHeight       int64
	FinalizedBlockHeight  int64
	TargetBlockHeight     int64
	ReadyTransactionCount int32
}

type LatestBlockMetrics struct {
	BestBlockHeight      int64
	FinalizedBlockHeight int64
}
