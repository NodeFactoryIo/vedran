package payout

import (
	gsrpc "github.com/NodeFactoryIo/go-substrate-rpc-client"
	"github.com/NodeFactoryIo/go-substrate-rpc-client/signature"
	"github.com/NodeFactoryIo/go-substrate-rpc-client/types"
	log "github.com/sirupsen/logrus"
	"math/big"
	"sync"
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
		nId := nodeId
		go func(to string, amount big.Int, wg *sync.WaitGroup, mux *sync.Mutex) {
			defer wg.Done()
			transactionDetails, err := executeTransaction(api, nId, to, amount, keyringPair, mux)
			if err != nil {
				fatalErrorsChannel <- err
			} else {
				resultsChannel <- transactionDetails
			}
		}(payoutDetails[nodeId].PayoutAddress, amount, &wg, &mux)
	}

	go func() {
		// wait for group To finish
		wg.Wait()
		close(wgDoneChannel)
		close(resultsChannel)
	}()

	var transactionDetails []*TransactionDetails
	var fatalErr error
	select {
	case <-wgDoneChannel:
		break
	case fatalErr = <-fatalErrorsChannel:
		break
	}
	// return even if just some of transaction have been executed
	for result := range resultsChannel {
		transactionDetails = append(transactionDetails, result)
	}
	return transactionDetails, fatalErr
}

func executeTransaction(
	api *gsrpc.SubstrateAPI,
	nodeId string,
	to string,
	amount big.Int,
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

	// listen for transaction Status
	defer sub.Unsubscribe()
	for {
		status := <-sub.Chan()
		if status.IsDropped {
			tx := &TransactionDetails{
				NodeId: nodeId,
				To:     to,
				Amount: amount,
				Status: Dropped,
			}
			log.Warningf("Dropped transaction: %v", tx)
			return tx, nil
		}
		if status.IsInvalid {
			tx := &TransactionDetails{
				NodeId: nodeId,
				To:     to,
				Amount: amount,
				Status: Invalid,
			}
			log.Warningf("Invalid transaction: %v", tx)
			return tx, nil
		}
		if status.IsFinalized {
			log.Debugf("Transaction for node %s completed at block hash: %#x\n", nodeId, status.AsFinalized)
			return &TransactionDetails{
				NodeId: nodeId,
				To:     to,
				Amount: amount,
				Status: Finalized,
			}, nil
		}
	}
}
