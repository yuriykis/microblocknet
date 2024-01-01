package store

import (
	"context"

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
	UTXOStore(context.Context) UTXOStorer
	TxStore(context.Context) TxStorer
	BlockStore(context.Context) BlockStorer
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

func (c *ChainMemoryStore) UTXOStore(ctx context.Context) UTXOStorer {
	if c.utxoStore == nil {
		c.utxoStore = NewMemoryUTXOStore()
	}
	return c.utxoStore
}

func (c *ChainMemoryStore) TxStore(ctx context.Context) TxStorer {
	if c.txStore == nil {
		c.txStore = NewMemoryTxStore()
	}
	return c.txStore
}

func (c *ChainMemoryStore) BlockStore(ctx context.Context) BlockStorer {
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

func (c *ChainMongoStore) UTXOStore(ctx context.Context) UTXOStorer {
	if c.utxoStore == nil {
		c.utxoStore = NewMongoUTXOStore(c.client)
	}
	return c.utxoStore
}

func (c *ChainMongoStore) TxStore(ctx context.Context) TxStorer {
	if c.txStore == nil {
		c.txStore = NewMongoTxStore(c.client)
	}
	return c.txStore
}

func (c *ChainMongoStore) BlockStore(ctx context.Context) BlockStorer {
	if c.blockStore == nil {
		c.blockStore = NewMongoBlockStore(c.client)
	}
	return c.blockStore
}

type TxStorer interface {
	Put(ctx context.Context, tx *proto.Transaction) error
	Get(ctx context.Context, txHash string) (*proto.Transaction, error)
	List(ctx context.Context) []*proto.Transaction
}

type BlockStorer interface {
	Put(ctx context.Context, block *proto.Block) error
	Get(ctx context.Context, blockHash string) (*proto.Block, error)
	List(ctx context.Context) []*proto.Block
}

type UTXOStorer interface {
	Put(ctx context.Context, utxo *proto.UTXO) error
	Get(ctx context.Context, key string) (*proto.UTXO, error)
	List(ctx context.Context) []*proto.UTXO
	GetByAddress(ctx context.Context, address []byte) ([]*proto.UTXO, error)
}
