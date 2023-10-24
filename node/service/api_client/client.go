package apiclient

import (
	"context"

	"github.com/yuriykis/microblocknet/node/service/types"
)

type Client interface {
	GetBlockByHeight(ctx context.Context, height int) (*types.GetBlockByHeightResponse, error)
	GetUTXOsByAddress(ctx context.Context, address []byte) (*types.GetUTXOsByAddressResponse, error)
	PeersAddrs(ctx context.Context) []string
}
