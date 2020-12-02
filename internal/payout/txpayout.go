package payout

import (
	gsrpc "github.com/NodeFactoryIo/go-substrate-rpc-client"
	"github.com/NodeFactoryIo/go-substrate-rpc-client/signature"
	"math/big"
	"sync"
)

func ExecuteAllPayoutTransactions(
	payoutDistribution map[string]big.Int,
	payoutDetails map[string]NodePayoutDetails,
	secret string,
	substrateRPCurl string,
) ([]*TransactionDetails, error) {
	api, err := gsrpc.NewSubstrateAPI(substrateRPCurl)
	if err != nil {
		return nil, err
	}

	keyringPair, err := signature.KeyringPairFromSecret(secret, "")
	if err != nil {
		return nil, err
	}

	return executeAllTransactions(payoutDistribution, payoutDetails, api, keyringPair)
}

func executeAllTransactions(
	payoutDistribution map[string]big.Int,
	payoutDetails map[string]NodePayoutDetails,
	api *gsrpc.SubstrateAPI,
	keyringPair signature.KeyringPair,
) ([]*TransactionDetails, error) {
	var mux sync.Mutex

	resultsChannel := make(chan *TransactionDetails, len(payoutDistribution))
	fatalErrorsChannel := make(chan error)
	waitGroupDoneChannel := make(chan bool, 1)

	var wg sync.WaitGroup
	// define number of goroutines
	wg.Add(len(payoutDistribution))

	for nodeId, amount := range payoutDistribution {
		nId := nodeId // create scoped variable to preserve value
		go func(to string, amount big.Int, wg *sync.WaitGroup, mux *sync.Mutex) {
			// execute transaction in separate goroutine and collect results in channels
			defer wg.Done()
			transactionDetails, err := ExecuteTransaction(api, nId, to, amount, keyringPair, mux)
			if err != nil {
				fatalErrorsChannel <- err
			} else {
				resultsChannel <- transactionDetails
			}
		}(payoutDetails[nodeId].PayoutAddress, amount, &wg, &mux)
	}

	go func() {
		// wait for group to finish
		wg.Wait()
		close(waitGroupDoneChannel)
		close(resultsChannel)
	}()

	return waitForTransactionDetails(waitGroupDoneChannel, fatalErrorsChannel, resultsChannel)
}

func waitForTransactionDetails(
	waitGroupDoneChannel chan bool,
	fatalErrorsChannel chan error,
	resultsChannel chan *TransactionDetails,
) ([]*TransactionDetails, error) {
	var transactionDetails []*TransactionDetails
	var fatalError error
	// wait for fatal error or all transactions executed
	select {
	case <-waitGroupDoneChannel:
		break
	case fatalError = <-fatalErrorsChannel:
		break
	}
	// return even if just some of transaction have been executed
	for result := range resultsChannel {
		transactionDetails = append(transactionDetails, result)
	}
	return transactionDetails, fatalError
}
