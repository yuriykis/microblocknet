package store

import (
	"github.com/yuriykis/microblocknet/proto"
)

type TxStorer interface {
	Put(tx *proto.Transaction) error
	Get(txID string) (*proto.Transaction, error)
}

type BlockStorer interface {
	Put(block *proto.Block) error
	Get(blockHash string) (*proto.Block, error)
}
