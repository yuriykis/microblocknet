package store

import (
	"fmt"
	"sync"

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

func (m *MemoryBlockStore) Put(block *proto.Block) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	hash := secure.HashBlock(block)
	m.blocks[hash] = block
	return nil
}

func (m *MemoryBlockStore) Get(blockHash string) (*proto.Block, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	block, ok := m.blocks[blockHash]
	if !ok {
		return nil, fmt.Errorf("block with id %s not found", blockHash)
	}
	return block, nil
}

func (m *MemoryBlockStore) List() []*proto.Block {
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

func (m *MongoBlockStore) Put(block *proto.Block) error {
	// TODO: implement
	return nil
}

func (m *MongoBlockStore) Get(blockID string) (*proto.Block, error) {
	// TODO: implement
	return nil, nil
}

func (m *MongoBlockStore) List() []*proto.Block {
	// TODO: implement
	return nil
}
