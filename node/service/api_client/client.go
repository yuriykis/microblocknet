package apiclient

import (
	"context"

	"github.com/yuriykis/microblocknet/common/requests"
)

type Client interface {
	GetBlockByHeight(ctx context.Context, height int) (*requests.GetBlockByHeightResponse, error)
	GetUTXOsByAddress(ctx context.Context, address []byte) (*requests.GetUTXOsByAddressResponse, error)
	PeersAddrs(ctx context.Context) []string
	NewTransaction(ctx context.Context, tReq requests.NewTransactionRequest) (requests.NewTransactionResponse, error)
}
