package requests

import "github.com/yuriykis/microblocknet/common/proto"

type InitTransactionRequest struct {
	FromAddress []byte
	FromPubKey  []byte
	ToAddress   []byte
	Amount      int
}

type InitTransactionResponse struct {
	Transaction *proto.Transaction
}

type NewTransactionRequest struct {
	Transaction *proto.Transaction
}

type NewTransactionResponse struct {
	Transaction *proto.Transaction
}

type GetMyUTXOsRequest struct {
	Address []byte
}

type GetMyUTXOsResponse struct {
	UTXOs []*proto.UTXO
}

type GetBlockByHeightRequest struct {
	Height int
}

type GetBlockByHeightResponse struct {
	Block *proto.Block
}

type GetUTXOsByAddressRequest struct {
	Address []byte
}

type GetUTXOsByAddressResponse struct {
	UTXOs []*proto.UTXO
}

type PeersAddrsRequest struct{}

type PeersAddrsResponse struct {
	PeersAddrs []string
}

type GetCurrentHeightResponse struct {
	Height int
}

type HealthcheckResponse struct {
	Healthcheck string
}

type RegisterNodeRequest struct {
	Address string
}
