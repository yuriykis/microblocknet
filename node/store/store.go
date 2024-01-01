package store

import (
	"github.com/sirupsen/logrus"
	"github.com/yuriykis/microblocknet/common/proto"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	storeTypeMemory = "memory"
	storeTypeMongo  = "mongo"

	mongoDBName = "microblocknet"
)

type Storer interface {
	UTXOStore() UTXOStorer
	TxStore() TxStorer
	BlockStore() BlockStorer
}

func NewChainStore(sType string) (Storer, error) {
	switch sType {
	case storeTypeMemory:
		return NewChainMemoryStore(), nil
	case storeTypeMongo:
		client, err := mongo.Connect(nil, nil)
		if err != nil {
			// TODO: adjust logger
			logrus.WithError(err).Error("failed to connect to mongo")
			return nil, err
		}
		return NewChainMongoStore(client), nil
	default:
		return NewChainMemoryStore(), nil
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

type ChainMongoStore struct {
	txStore    TxStorer
	blockStore BlockStorer
	utxoStore  UTXOStorer

	client *mongo.Client
}

func NewChainMongoStore(client *mongo.Client) Storer {
	return &ChainMongoStore{
		client: client,
	}
}

func (c *ChainMongoStore) UTXOStore() UTXOStorer {
	if c.utxoStore == nil {
		c.utxoStore = NewMongoUTXOStore(c.client)
	}
	return c.utxoStore
}

func (c *ChainMongoStore) TxStore() TxStorer {
	if c.txStore == nil {
		c.txStore = NewMongoTxStore(c.client)
	}
	return c.txStore
}

func (c *ChainMongoStore) BlockStore() BlockStorer {
	if c.blockStore == nil {
		c.blockStore = NewMongoBlockStore(c.client)
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
