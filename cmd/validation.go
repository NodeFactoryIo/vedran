package cmd

import (
	"errors"
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/ui/prompts"
	"strconv"
)

func ValidatePayoutFlags(
	payoutReward string,
	payoutAddress string,
	showPrompts bool,
	) (float64, error) {
	var err error
	var rewardAsFloat64 float64

	// if total reward is determined as wallet balance
	if payoutReward == "-1" {
		if payoutAddress == "" {
			return 0, errors.New("Unable to set reward amount to entire wallet balance if fee address not provided")
		} else {
			if showPrompts {
				confirmed, err := prompts.ShowConfirmationPrompt(
					fmt.Sprintf("You choose that reward amount is defined as entire balance on lb wallet!" +
						"On payout entire balance will be distributed as reward and lb fee will be sent to address %s",
						payoutAddress),
				)
				if err != nil {
					return 0, err
				}
				if !confirmed {
					return 0, errors.New("Payout configuration canceled")
				}
			}
		}
	} else {
		rewardAsFloat64, err = strconv.ParseFloat(payoutReward, 64)
		if err != nil || rewardAsFloat64 < -1 {
			return 0, errors.New("invalid total reward value")
		}
	}

	return rewardAsFloat64, nil
}
