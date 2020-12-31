package api

import (
	gsrpc "github.com/NodeFactoryIo/go-substrate-rpc-client"
	"github.com/NodeFactoryIo/go-substrate-rpc-client/types"
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
