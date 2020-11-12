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

func CalculatePayoutDistributionByNode(
	payoutDetails map[string]models.NodeStatsDetails,
	totalReward float64,
	loadBalancerFeePercentage float64,
) map[string]big.Int {
	var rewardPool = totalReward

	loadbalancerReward := rewardPool * loadBalancerFeePercentage
	rewardPool -= loadbalancerReward

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
	payoutAmountDistributionByNodes := make(map[string]big.Int, len(payoutDetails))

	for nodeId, nodeStatsDetails := range payoutDetails {
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
		payoutAmountDistributionByNodes[nodeId] = *totalNodeRewardAsInt
	}

	return payoutAmountDistributionByNodes
}
