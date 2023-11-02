package apiclient

import (
	"context"

	"github.com/yuriykis/microblocknet/common/requests"
)

type Client interface {
	String() string
	Healthcheck(ctx context.Context) (requests.HealthcheckResponse, error)
	GetBlockByHeight(ctx context.Context, height int) (requests.GetBlockByHeightResponse, error)
	GetUTXOsByAddress(ctx context.Context, address []byte) (*requests.GetUTXOsByAddressResponse, error)
	PeersAddrs(ctx context.Context) []string
	NewTransaction(ctx context.Context, tReq requests.NewTransactionRequest) (requests.NewTransactionResponse, error)
	Height(ctx context.Context) (requests.GetCurrentHeightResponse, error)
}
