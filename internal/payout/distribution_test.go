package payout

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func Test_CalculatePayoutDistributionByNode(t *testing.T) {
	tests := []struct {
		name               string
		payoutDetails      map[string]models.NodeStatsDetails
		totalReward        float64
		loadBalancerFee    float64
		resultDistribution map[string]big.Int
	}{
		{   // this test is set for 10/90 split between liveliness and requests
			name: "test distribution",
			payoutDetails: map[string]models.NodeStatsDetails{
				"1": {
					TotalPings:    100,
					TotalRequests: 10,
				},
				"2": {
					TotalPings:    100,
					TotalRequests: 5,
				},
				"3": {
					TotalPings:    90,
					TotalRequests: 10,
				},
				"4": {
					TotalPings:    90,
					TotalRequests: 5,
				},
				"5": {
					TotalPings:    50,
					TotalRequests: 2,
				},
				"6": {
					TotalPings:    40,
					TotalRequests: 0,
				},
			},
			totalReward: 100000000,
			loadBalancerFee: 0.1,
			resultDistribution: map[string]big.Int{
				"1": *big.NewInt(27227393), // 27227393.617021276 // 100P 10R
				"2": *big.NewInt(14571143), // 14571143.617021276 // 100P 5R
				"3": *big.NewInt(27035904), // 27035904.255319147 // 90P  10R
				"4": *big.NewInt(14379654), // 14379654.25531915  // 90P  5R
				"5": *big.NewInt(6019946),  // 6019946.808510638  // 50P  2R
				"6": *big.NewInt(765957),   // 765957.4468085106  // 40P  0R
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			distributionByNode := CalculatePayoutDistributionByNode(
				test.payoutDetails, test.totalReward, test.loadBalancerFee,
			)
			assert.Equal(t, test.resultDistribution, distributionByNode)
			totalDistributed := big.NewInt(0)
			for _, amount := range distributionByNode {
				totalDistributed.Add(totalDistributed, &amount)
			}
			totalShoudBeDistributed := test.totalReward * (float64(1) - test.loadBalancerFee)

			totalShouldBeDistributedRounded, _ := big.NewFloat(totalShoudBeDistributed).Int(nil)
			delta := big.NewInt(0).Sub(totalShouldBeDistributedRounded, totalDistributed)
			assert.GreaterOrEqual(t, delta.Int64(), int64(0))
		})
	}
}
