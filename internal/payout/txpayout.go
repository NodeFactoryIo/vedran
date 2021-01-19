package payout

import (
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v2"
	"github.com/centrifuge/go-substrate-rpc-client/v2/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
	"github.com/pkg/errors"
	"math/big"
	"sync"
)

func ExecuteAllPayoutTransactions(
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

	metadataLatest, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get latest metadat")
	}

	nonce, err := GetNonce(metadataLatest, keyringPair, api)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get nonce")
	}

	for nodePayoutAddress, amount := range payoutDistribution {
		// execute transaction in separate goroutine and collect results in channels
		go func(to string, amount big.Int, wg *sync.WaitGroup, mux *sync.Mutex, nonce uint32) {
			defer wg.Done()
			transactionDetails, err := ExecuteTransaction(api, to, amount, keyringPair, mux, metadataLatest, nonce)
			if err != nil {
				fatalErrorsChannel <- err
			} else {
				resultsChannel <- transactionDetails
			}
		}(nodePayoutAddress, amount, &wg, &mux, nonce)
		nonce += 1
	}

	go func() {
		// wait for group to finish
		wg.Wait()
		close(waitGroupDoneChannel)
		close(resultsChannel)
	}()

	return waitForTransactionDetails(waitGroupDoneChannel, fatalErrorsChannel, resultsChannel)
}

func GetNonce(metadataLatest *types.Metadata, keyringPair signature.KeyringPair, api *gsrpc.SubstrateAPI) (uint32, error) {
	storageKey, err := types.CreateStorageKey(
		metadataLatest,
		"System",
		"Account",
		keyringPair.PublicKey,
		nil,
	)
	if err != nil {
		return 0, err
	}

	var accountInfo types.AccountInfo
	ok, err := api.RPC.State.GetStorageLatest(storageKey, &accountInfo)
	if err != nil || !ok {
		return 0, err
	}

	nonce := uint32(accountInfo.Nonce)
	return nonce, err
}

func GetBalance(metadataLatest *types.Metadata, keyringPair signature.KeyringPair, api *gsrpc.SubstrateAPI) (types.U128, error) {
	address, err := types.HexDecodeString(keyringPair.Address)
	if err != nil {
		return types.U128{}, err
	}

	key, err := types.CreateStorageKey(metadataLatest, "Balances", "FreeBalance", address, nil)
	if err != nil {
		return types.U128{}, err
	}

	// Retrieve the initial balance
	var balance types.U128
	ok, err := api.RPC.State.GetStorageLatest(key, &balance)
	if err != nil || !ok {
		return types.U128{}, err
	}

	return balance, nil
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
