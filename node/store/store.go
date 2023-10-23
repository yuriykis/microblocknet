package store

import (
	"github.com/yuriykis/microblocknet/node/proto"
)

type TxStorer interface {
	Put(tx *proto.Transaction) error
	Get(txHash string) (*proto.Transaction, error)
	List() []*proto.Transaction
}

type BlockStorer interface {
	Put(block *proto.Block) error
	Get(blockHash string) (*proto.Block, error)
	List() []*proto.Block
}

type UTXOStorer interface {
	Put(utxo *proto.UTXO) error
	Get(key string) (*proto.UTXO, error)
	List() []*proto.UTXO
	GetByAddress(address []byte) ([]*proto.UTXO, error)
}
