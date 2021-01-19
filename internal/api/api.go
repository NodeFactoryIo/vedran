package api

import (
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v2"
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
)

func InitializeSubstrateAPI(substrateRPCUrl string) (*gsrpc.SubstrateAPI, error) {
	api, err := gsrpc.NewSubstrateAPI(substrateRPCUrl)
	if err != nil {
		return nil, err
	}

	opts := types.SerDeOptions{NoPalletIndices: true}
	types.SetSerDeOptions(opts)

	return api, nil
}
