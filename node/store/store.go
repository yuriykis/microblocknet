package store

import (
	"github.com/yuriykis/microblocknet/common/proto"
)

const (
	storeTypeMemory = "memory"
)

type Storer interface {
	UTXOStore() UTXOStorer
	TxStore() TxStorer
	BlockStore() BlockStorer
}

func NewChainStore(sType string) Storer {
	switch sType {
	case storeTypeMemory:
		return NewChainMemoryStore()
	default:
		return NewChainMemoryStore()
	}
}

func NewChainMemoryStore() Storer {
	return &ChainMemoryStore{}
}

type ChainMemoryStore struct {
	txStore    TxStorer
	blockStore BlockStorer
	utxoStore  UTXOStorer
}

func (c *ChainMemoryStore) UTXOStore() UTXOStorer {
	if c.utxoStore == nil {
		c.utxoStore = NewMemoryUTXOStore()
	}
	return c.utxoStore
}

func (c *ChainMemoryStore) TxStore() TxStorer {
	if c.txStore == nil {
		c.txStore = NewMemoryTxStore()
	}
	return c.txStore
}

func (c *ChainMemoryStore) BlockStore() BlockStorer {
	if c.blockStore == nil {
		c.blockStore = NewMemoryBlockStore()
	}
	return c.blockStore
}

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
