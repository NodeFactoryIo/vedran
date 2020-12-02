package payout

import (
	gsrpc "github.com/NodeFactoryIo/go-substrate-rpc-client"
	"github.com/NodeFactoryIo/go-substrate-rpc-client/signature"
	"github.com/NodeFactoryIo/go-substrate-rpc-client/types"
	"math/big"
	"sync"
)

func ExecuteAllPayoutTransactions(
	payoutDistribution map[string]big.Int,
	secret string,
	substrateRPCurl string,
) ([]*TransactionDetails, error) {
	api, err := gsrpc.NewSubstrateAPI(substrateRPCurl)
	if err != nil {
		return nil, err
	}

	opts := types.SerDeOptions{NoPalletIndices: true}
	types.SetSerDeOptions(opts)

	keyringPair, err := signature.KeyringPairFromSecret(secret, "")
	if err != nil {
		return nil, err
	}

	return executeAllTransactions(payoutDistribution, api, keyringPair)
}

func executeAllTransactions(
	payoutDistribution map[string]big.Int,
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

	for nodePayoutAddress, amount := range payoutDistribution {
		// execute transaction in separate goroutine and collect results in channels
		go func(to string, amount big.Int, wg *sync.WaitGroup, mux *sync.Mutex) {
			defer wg.Done()
			transactionDetails, err := ExecuteTransaction(api, to, amount, keyringPair, mux)
			if err != nil {
				fatalErrorsChannel <- err
			} else {
				resultsChannel <- transactionDetails
			}
		}(nodePayoutAddress, amount, &wg, &mux)
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
