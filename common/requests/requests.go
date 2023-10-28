package requests

import "github.com/yuriykis/microblocknet/common/proto"

type CreateTransactionRequest struct {
	FromAddress []byte
	ToAddress   []byte
	Amount      int
}

type CreateTransactionResponse struct {
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
