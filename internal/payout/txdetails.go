package payout

import (
	"fmt"
	"github.com/NodeFactoryIo/go-substrate-rpc-client/rpc/author"
	log "github.com/sirupsen/logrus"
	"math/big"
)

type TransactionStatus string

const (
	Finalized = TransactionStatus("Finalized")
	Dropped   = TransactionStatus("Dropped")
	Invalid   = TransactionStatus("Invalid")
)

type TransactionDetails struct {
	NodeId string
	To     string
	Amount big.Int
	Status TransactionStatus
}

func listenForTransactionStatus(
	sub *author.ExtrinsicStatusSubscription,
	transactionDetails TransactionDetails,
) TransactionDetails {
	defer sub.Unsubscribe()
	for {
		status := <-sub.Chan()
		if status.IsDropped {
			transactionDetails.Status = Dropped
			log.Warningf("Dropped transaction: %v", transactionDetails)
			return transactionDetails
		}
		if status.IsInvalid {
			transactionDetails.Status = Invalid
			log.Warningf("Invalid transaction: %v", transactionDetails)
			return transactionDetails
		}
		if status.IsFinalized {
			transactionDetails.Status = Finalized
			log.Debugf(
				"Transaction for node %s completed at block hash: %#x\n",
				transactionDetails.NodeId,
				status.AsFinalized,
			)
			return transactionDetails
		}
		fmt.Println(status)
	}
}
