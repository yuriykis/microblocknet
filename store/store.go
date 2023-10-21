package store

import (
	"github.com/yuriykis/microblocknet/proto"
)

type TxStorer interface {
	Put(tx *proto.Transaction) error
	Get(txHash string) (*proto.Transaction, error)
}

type BlockStorer interface {
	Put(block *proto.Block) error
	Get(blockHash string) (*proto.Block, error)
}

type UTXOStorer interface {
	Put(utxo *proto.UTXO) error
	Get(key string) (*proto.UTXO, error)
	GetByAddress(address []byte) ([]*proto.UTXO, error)
}
