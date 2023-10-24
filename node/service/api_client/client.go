package apiclient

import "github.com/yuriykis/microblocknet/node/proto"

type Client interface {
	GetBlockByHeight(height int) (*proto.Block, error)
	GetUTXOsByAddress(address []byte) ([]*proto.UTXO, error)
	PeersAddrs() []string
}
