package payout

import (
	gsrpc "github.com/centrifuge/go-substrate-rpc-client"
	"github.com/centrifuge/go-substrate-rpc-client/signature"
	"github.com/centrifuge/go-substrate-rpc-client/types"
	log "github.com/sirupsen/logrus"
	"math/big"
	"sync"
)

type TransactionStatus string

const (
	Finalized = TransactionStatus("Finalized")
	Dropped   = TransactionStatus("Dropped")
)

type TransactionDetails struct {
	to     string
	amount big.Int
	status TransactionStatus
}

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

	var wg sync.WaitGroup
	var mux sync.Mutex

	resultsChannel := make(chan *TransactionDetails, len(payoutDistribution))
	fatalErrorsChannel := make(chan error)
	wgDoneChannel := make(chan bool, 1)

	wg.Add(len(payoutDistribution)) // define number of goroutines
	for nodeId, amount := range payoutDistribution {
		go func(to string, amount big.Int, wg *sync.WaitGroup, mux *sync.Mutex) {
			defer wg.Done()
			transactionDetails, err := executeTransaction(api, to, amount, keyringPair, mux)
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
		close(wgDoneChannel)
		close(resultsChannel)
	}()

	var transactionDetails []*TransactionDetails
	select {
	case <-wgDoneChannel:
		break
	case err := <-fatalErrorsChannel:
		// return if some of transaction have been executed
		for result := range resultsChannel {
			transactionDetails = append(transactionDetails, result)
		}
		return transactionDetails, err
	}

	for result := range resultsChannel {
		transactionDetails = append(transactionDetails, result)
	}
	return transactionDetails, nil
}

func executeTransaction(
	api *gsrpc.SubstrateAPI,
	to string, amount big.Int,
	keyringPair signature.KeyringPair,
	mux *sync.Mutex,
) (*TransactionDetails, error) {

	// lock segment so goroutines don't access api at the same time
	mux.Lock()

	metadataLatest, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return nil, err
	}

	toAddress, err := types.NewAddressFromHexAccountID(to)
	if err != nil {
		return nil, err
	}

	call, err := types.NewCall(
		metadataLatest,
		"Balances.transfer",
		toAddress,
		types.NewUCompact(&amount),
	)
	if err != nil {
		return nil, err
	}

	extrinsic := types.NewExtrinsic(call)

	genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		return nil, err
	}

	runtimeVersionLatest, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		return nil, err
	}

	storageKey, err := types.CreateStorageKey(
		metadataLatest,
		"System",
		"AccountNonce",
		keyringPair.PublicKey,
		nil,
	)
	if err != nil {
		return nil, err
	}

	var nonce uint32
	_, err = api.RPC.State.GetStorageLatest(storageKey, nonce)
	if err != nil {
		return nil, err
	}

	signatureOptions := types.SignatureOptions{
		Era:         types.ExtrinsicEra{IsMortalEra: false},
		Nonce:       types.NewUCompactFromUInt(uint64(nonce)),
		Tip:         types.NewUCompactFromUInt(0),
		SpecVersion: runtimeVersionLatest.SpecVersion,
		GenesisHash: genesisHash,
		BlockHash:   genesisHash,
	}

	err = extrinsic.Sign(keyringPair, signatureOptions)
	if err != nil {
		return nil, err
	}

	sub, err := api.RPC.Author.SubmitAndWatchExtrinsic(extrinsic)
	if err != nil {
		return nil, err
	}

	// unlock segment
	mux.Unlock()

	defer sub.Unsubscribe()
	for {
		status := <-sub.Chan()
		if status.IsDropped {
			log.Debug("Dropped transaction")
			return &TransactionDetails{
				to:     to,
				amount: amount,
				status: Dropped,
			}, nil
		}
		if status.IsFinalized {
			log.Debugf("Completed at block hash: %#x\n", status.AsFinalized)
			return &TransactionDetails{
				to:     to,
				amount: amount,
				status: Finalized,
			}, nil
		}
		log.Debugf("Transaction status: %v#\n", status)
	}
}
