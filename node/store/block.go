package store

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/node/secure"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	blockColl = "block"
)

type MemoryBlockStore struct {
	lock   sync.RWMutex
	blocks map[string]*proto.Block
}

func NewMemoryBlockStore() *MemoryBlockStore {
	return &MemoryBlockStore{
		blocks: make(map[string]*proto.Block),
	}
}

func (m *MemoryBlockStore) Put(ctx context.Context, block *proto.Block) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	hash := secure.HashBlock(block)
	m.blocks[hash] = block
	return nil
}

func (m *MemoryBlockStore) Get(ctx context.Context, blockHash string) (*proto.Block, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	block, ok := m.blocks[blockHash]
	if !ok {
		return nil, fmt.Errorf("block with id %s not found", blockHash)
	}
	return block, nil
}

func (m *MemoryBlockStore) List(ctx context.Context) []*proto.Block {
	m.lock.RLock()
	defer m.lock.RUnlock()

	blocks := make([]*proto.Block, 0)
	for _, block := range m.blocks {
		blocks = append(blocks, block)
	}
	return blocks
}

// -----------------------------------------------------------------------------

type MongoBlockStore struct {
	client *mongo.Client
	coll   *mongo.Collection
}

func NewMongoBlockStore(client *mongo.Client) *MongoBlockStore {
	return &MongoBlockStore{
		client: client,
		coll:   client.Database(mongoDBName).Collection(blockColl),
	}
}

func (m *MongoBlockStore) Put(ctx context.Context, block *proto.Block) error {
	res, err := m.coll.InsertOne(ctx, block)
	if err != nil {
		return err
	}
	logrus.Infof("inserted block %s", res.InsertedID)
	return nil
}

func (m *MongoBlockStore) Get(ctx context.Context, blockID string) (*proto.Block, error) {
	var block proto.Block
	if err := m.coll.FindOne(ctx, blockID).Decode(&block); err != nil {
		return nil, err
	}
	return &block, nil
}

func (m *MongoBlockStore) List(ctx context.Context) []*proto.Block {
	var blocks []*proto.Block
	cur, err := m.coll.Find(ctx, nil)
	if err != nil {
		return nil
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var block proto.Block
		if err := cur.Decode(&block); err != nil {
			return nil
		}
		blocks = append(blocks, &block)
	}
	return blocks
}
