package stats

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"math"
	"time"
)

func CalculateTotalPingsForNode(
	repos repositories.Repos,
	nodeId string,
	intervalStart time.Time,
	intervalEnd time.Time,
) (float64, error) {
	downtimesInInterval, err := repos.DowntimeRepo.FindDowntimesInsideInterval(nodeId, intervalStart, intervalEnd)
	if err != nil {
		if err.Error() == "not found" {
			downtimesInInterval = []models.Downtime{}
		} else {
			return 0, err
		}
	}

	totalTime := intervalEnd.Sub(intervalStart)
	leftTime := totalTime
	for _, downtime := range downtimesInInterval {
		var downtimeLength time.Duration
		// case 1: entire downtime inside interval
		if downtime.Start.After(intervalStart) && downtime.End.Before(intervalEnd) {
			downtimeLength = downtime.End.Sub(downtime.Start)
		}
		// case 2: downtime started before interval
		if downtime.Start.Before(intervalStart) {
			downtimeLength = downtime.End.Sub(intervalStart)
		}
		leftTime -= downtimeLength
	}
	// case 3: downtime still active
	_, duration, err := repos.PingRepo.CalculateDowntime(nodeId, intervalEnd)
	if err != nil {
		return 0, err
	}
	if duration.Seconds() > leftTime.Seconds() {
		// if node was down for entire observed interval
		return 0, nil
	}
	if math.Abs(duration.Seconds()) > pingIntervalSeconds {
		leftTime -= duration
	}

	totalPings := leftTime.Seconds() / pingIntervalSeconds
	return totalPings, nil
}