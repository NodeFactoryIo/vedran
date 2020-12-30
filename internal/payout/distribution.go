package payout

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"math"
	"math/big"
)

const (
	livelinessRewardPercentage = 0.1
	requestsRewardPercentage   = 0.9
)

type LoadBalancerDistributionConfiguration struct {
	FeePercentage       float64
	PayoutAddress       string
	DifferentFeeAddress bool
}

func CalculatePayoutDistributionByNode(
	payoutDetails map[string]models.NodeStatsDetails,
	totalReward float64,
	lbConfiguration LoadBalancerDistributionConfiguration,
) map[string]big.Int {
	var rewardPool = totalReward
	numOfNodes := len(payoutDetails)
	if lbConfiguration.DifferentFeeAddress {
		// lb has separate address for lb fee
		numOfNodes += 1
	}
	payoutAmountDistributionByNodes := make(map[string]big.Int, numOfNodes)

	loadbalancerReward := rewardPool * lbConfiguration.FeePercentage
	rewardPool -= loadbalancerReward
	if lbConfiguration.DifferentFeeAddress {
		lbRewardAsInt, _ := big.NewFloat(loadbalancerReward).Int(nil)
		payoutAmountDistributionByNodes[lbConfiguration.PayoutAddress] = *lbRewardAsInt
	}

	livelinessRewardPool := rewardPool * livelinessRewardPercentage
	requestsRewardPool := rewardPool * requestsRewardPercentage

	var totalNumberOfPings = float64(0)
	var totalNumberOfRequests = float64(0)
	for _, node := range payoutDetails {
		totalNumberOfPings += node.TotalPings
		totalNumberOfRequests += node.TotalRequests
	}

	totalDistributedLivelinessRewards := float64(0)
	totalDistributedRequestsRewards := float64(0)

	for nodeAddress, nodeStatsDetails := range payoutDetails {
		// liveliness rewards
		livelinessReward := float64(0)
		if totalNumberOfPings != 0 && nodeStatsDetails.TotalPings != 0 {
			nodeLivelinessRewardPercentage := nodeStatsDetails.TotalPings / totalNumberOfPings
			livelinessReward = livelinessRewardPool * nodeLivelinessRewardPercentage
			livelinessReward = math.Floor(livelinessReward)
			totalDistributedLivelinessRewards += livelinessReward
		}
		// requests rewards
		requestsReward := float64(0)
		if totalNumberOfRequests != 0 && nodeStatsDetails.TotalRequests != 0 {
			nodeRequestsRewardPercentage := nodeStatsDetails.TotalRequests / totalNumberOfRequests
			requestsReward = requestsRewardPool * nodeRequestsRewardPercentage
			requestsReward = math.Floor(requestsReward)
			totalDistributedRequestsRewards += requestsReward
		}

		totalNodeReward := livelinessReward + requestsReward
		totalNodeRewardAsInt, _ := big.NewFloat(totalNodeReward).Int(nil)
		payoutAmountDistributionByNodes[nodeAddress] = *totalNodeRewardAsInt
	}

	return payoutAmountDistributionByNodes
}
