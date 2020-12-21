package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidatePayoutFlags(t *testing.T) {
	tests := []struct {
		name string
		payoutReward string
		payoutAddress string
		validateReturns float64
		validateError bool
	}{
		{
			name: "valid flags",
			payoutReward: "1000",
			payoutAddress: "",
			validateReturns: float64(1000),
			validateError: false,
		},
		{
			name: "invalid flags, missing reward address",
			payoutReward: "-1",
			payoutAddress: "",
			validateReturns: float64(0),
			validateError: true,
		},
		{
			name: "invalid flags, negative reward",
			payoutReward: "-100",
			payoutAddress: "",
			validateReturns: float64(0),
			validateError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rewAsFloat64, err := ValidatePayoutFlags(test.payoutReward, test.payoutAddress, false)
			assert.Equal(t, test.validateReturns, rewAsFloat64)
			if test.validateError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
