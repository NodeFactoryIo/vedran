package prometheus

import (
	"runtime"
	"strconv"
	"time"

	"github.com/NodeFactoryIo/vedran/internal/configuration"
	"github.com/NodeFactoryIo/vedran/internal/payout"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"github.com/NodeFactoryIo/vedran/pkg/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
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
	payoutDistribution = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "vedran_payout_distribution",
			Help: "Payout distribution per polkadot address",
		},
		[]string{"address"},
	)
)

// RecordMetrics starts goroutines for recording metrics
func RecordMetrics(repos repositories.Repos) {
	version := promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "vedran_version",
			Help: "App and golang version of vedran",
			ConstLabels: map[string]string{
				"go_version":     runtime.Version(),
				"vedran_version": version.Version,
			},
		},
	)
	version.Set(1)

	go recordPayoutDistribution(repos)
	go recordActiveNodeCount(repos.NodeRepo)
	go recordPenalizedNodeCount(repos.NodeRepo)
	go recordSuccessfulRequestCount(repos.RecordRepo)
	go recordFailedRequestCount(repos.RecordRepo)
}

func recordPayoutDistribution(repos repositories.Repos) {
	for {
		stats, err := payout.GetStatsForPayout(repos, time.Now(), false)
		if err != nil {
			log.Errorf("Failed recording stats for payout because of: %v", err)
			time.Sleep(15 * time.Minute)
			continue
		}
		distributionByNode := payout.CalculatePayoutDistributionByNode(
			stats,
			100,
			float64(configuration.Config.Fee),
		)

		for address, distribution := range distributionByNode {
			floatDistribution, _ := strconv.ParseFloat(distribution.String(), 64)
			payoutDistribution.With(
				prometheus.Labels{"address": address},
			).Set(
				floatDistribution,
			)
		}

		time.Sleep(1 * time.Minute)
	}
}

func recordActiveNodeCount(nodeRepo repositories.NodeRepository) {
	for {
		activeNodes.Set(float64(len(*nodeRepo.GetAllActiveNodes())))
		time.Sleep(15 * time.Second)
	}
}

func recordPenalizedNodeCount(nodeRepo repositories.NodeRepository) {
	for {
		nodes, _ := nodeRepo.GetPenalizedNodes()
		penalizedNodes.Set(float64(len(*nodes)))
		time.Sleep(15 * time.Second)
	}
}

func recordSuccessfulRequestCount(recordRepo repositories.RecordRepository) {
	for {
		count, _ := recordRepo.CountSuccessfulRequests()
		successfulRequests.Set(float64(count))
		time.Sleep(15 * time.Second)
	}
}

func recordFailedRequestCount(recordRepo repositories.RecordRepository) {
	for {
		count, _ := recordRepo.CountFailedRequests()
		failedRequests.Set(float64(count))
		time.Sleep(15 * time.Second)
	}
}
