package payout

import (
	gsrpc "github.com/NodeFactoryIo/go-substrate-rpc-client"
	"github.com/NodeFactoryIo/go-substrate-rpc-client/signature"
	"github.com/NodeFactoryIo/go-substrate-rpc-client/types"
	"github.com/decred/base58"
	"math/big"
	"sync"
)

func ExecuteTransaction(
	api *gsrpc.SubstrateAPI,
	to string,
	amount big.Int,
	keyringPair signature.KeyringPair,
	mux *sync.Mutex,
	metadataLatest *types.Metadata,
	nonce uint32,
) (*TransactionDetails, error) {
	// lock segment so goroutines don't access api at the same time
	mux.Lock()

	decoded := base58.Decode(to)
	// remove the 1st byte (network identifier) & last 2 bytes (blake2b hash)
	pubKey := decoded[1 : len(decoded)-2]
	toAddress := types.NewAddressFromAccountID(pubKey)

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

	signatureOptions := types.SignatureOptions{
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		Nonce:              types.NewUCompactFromUInt(uint64(nonce)),
		Tip:                types.NewUCompactFromUInt(0),
		SpecVersion:        runtimeVersionLatest.SpecVersion,
		GenesisHash:        genesisHash,
		BlockHash:          genesisHash,
		TransactionVersion: runtimeVersionLatest.TransactionVersion,
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

	txDetails := listenForTransactionStatus(
		sub,
		TransactionDetails{
			To:     to,
			Amount: amount,
		},
	)
	return &txDetails, nil
}
