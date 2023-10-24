package types

import "github.com/yuriykis/microblocknet/node/proto"

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
