package prometheus

import (
	"runtime"
	"time"

	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	activeNodes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "vedran_number_of_active_nodes",
		Help: "The total number of active nodes serving requests",
	})
	penalizedNodes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "vedran_number_of_penalized_nodes",
		Help: "The total number of nodes which are on cooldown",
	})
	successfulRequests = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "vedran_number_of_successful_requests",
		Help: "The total number of successful requests served via vedran",
	})
	failedRequests = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "vedran_number_of_failed_requests",
		Help: "The total number of successful requests served via vedran",
	})
)

func RecordMetrics(repos repositories.Repos) {
	version := promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "vedran_version",
			Help: "App and golang version of vedran",
			ConstLabels: map[string]string{
				"go_version":     runtime.Version(),
				"vedran_version": "1.13.1",
			},
		},
	)
	version.Set(1)

	go func() {
		for {
			activeNodes.Set(float64(len(*repos.NodeRepo.GetAllActiveNodes())))
			time.Sleep(15 * time.Second)
		}
	}()
	go func() {
		for {
			nodes, _ := repos.NodeRepo.GetPenalizedNodes()
			penalizedNodes.Set(float64(len(*nodes)))
			time.Sleep(15 * time.Second)
		}
	}()
	go func() {
		for {
			count, _ := repos.RecordRepo.CountSuccessfulRequests()
			successfulRequests.Set(float64(count))
			time.Sleep(15 * time.Second)
		}
	}()
	go func() {
		for {
			count, _ := repos.RecordRepo.CountFailedRequests()
			failedRequests.Set(float64(count))
			time.Sleep(15 * time.Second)
		}
	}()
}
