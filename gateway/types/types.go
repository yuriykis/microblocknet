package types

import "github.com/yuriykis/microblocknet/node/proto"

type CreateTransactionRequest struct {
	FromAddress []byte
	ToAddress   []byte
	Amount      int
}

type CreateTransactionResponse struct {
	TransactionHash []byte
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
