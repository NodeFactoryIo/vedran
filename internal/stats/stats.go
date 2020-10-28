package stats

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"math"
	"time"
)

const (
	pingIntervalSeconds = 10
)

func CalculateStatisticsForPayout(repos repositories.Repos) (map[string]models.NodePaymentDetails, error) {
	allNodes, err := repos.NodeRepo.GetAll()
	if err != nil {
		// todo
		return nil, err
	}

	IntervalStart := time.Now().Add(-48 * time.Hour)
	IntervalEnd := time.Now()

	var allNodesStats = make(map[string]models.NodePaymentDetails)
	for _, node := range *allNodes {
		recordsInInterval, err := repos.RecordRepo.FindSuccessfulRecordsInsideInterval(node.ID, IntervalStart, IntervalEnd)
		if err != nil {
			if err.Error() == "not found" {
				recordsInInterval = []models.Record{}
			} else {
				return nil, err
			}
		}
		downtimesInInterval, err := repos.DowntimeRepo.FindDowntimesInsideInterval(node.ID, IntervalStart, IntervalEnd)
		if err != nil {
			if err.Error() == "not found" {
				downtimesInInterval = []models.Downtime{}
			} else {
				return nil, err
			}
		}

		totalTime := IntervalEnd.Sub(IntervalStart)
		leftTime := totalTime
		for _, downtime := range downtimesInInterval {
			var downtimeLength time.Duration
			// case 1: whole downtime inside interval
			if downtime.Start.After(IntervalStart) && downtime.End.Before(IntervalEnd) {
				downtimeLength = downtime.End.Sub(downtime.Start)
			}
			// case 2: downtime started before interval
			if downtime.Start.Before(IntervalStart) {
				downtimeLength = downtime.End.Sub(IntervalStart)
			}
			if downtimeLength == time.Duration(0) {
				// todo
			}
			leftTime -= downtimeLength
		}
		// case 3: downtime still active
		_, duration, err := repos.PingRepo.CalculateDowntime(node.ID, IntervalEnd)
		if err != nil {
			return nil, err // todo
		}
		if math.Abs(duration.Seconds()) > pingIntervalSeconds { // todo
			leftTime -= duration
		}

		totalPings := leftTime.Seconds() / pingIntervalSeconds

		allNodesStats[node.ID] = models.NodePaymentDetails{
			TotalPings:    totalPings,
			TotalRequests: float64(len(recordsInInterval)),
		}
	}

	return allNodesStats, nil
}