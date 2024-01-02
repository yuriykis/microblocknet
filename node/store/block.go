package store

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/node/secure"
	"go.mongodb.org/mongo-driver/bson"
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

// Put inserts a block into the database, implementing the BlockStorer interface.
func (m *MongoBlockStore) Put(ctx context.Context, block *proto.Block) error {
	hash := secure.HashBlock(block)
	res, err := m.coll.InsertOne(ctx, bson.M{
		"hash":  hex.EncodeToString([]byte(hash)),
		"block": block,
	})
	if err != nil {
		return err
	}
	logrus.Infof("inserted block %s", res.InsertedID)
	return nil
}

// Get retrieves a block from the database, implementing the BlockStorer interface.
func (m *MongoBlockStore) Get(ctx context.Context, blockID string) (*proto.Block, error) {
	var blockDoc struct {
		Hash  string       `bson:"hash"`
		Block *proto.Block `bson:"block"`
	}
	if err := m.coll.FindOne(ctx, blockID).Decode(&blockDoc); err != nil {
		return nil, err
	}
	return blockDoc.Block, nil
}

// List retrieves all blocks from the database, implementing the BlockStorer interface.
func (m *MongoBlockStore) List(ctx context.Context) []*proto.Block {
	blocksDocs := make([]struct {
		Hash  string       `bson:"hash"`
		Block *proto.Block `bson:"block"`
	}, 0)
	cur, err := m.coll.Find(ctx, nil)
	if err != nil {
		return nil
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var blockDoc struct {
			Hash  string       `bson:"hash"`
			Block *proto.Block `bson:"block"`
		}
		if err := cur.Decode(&blockDoc); err != nil {
			return nil
		}
		blocksDocs = append(blocksDocs, blockDoc)
	}
	blocks := make([]*proto.Block, len(blocksDocs))
	for i, blockDoc := range blocksDocs {
		blocks[i] = blockDoc.Block
	}
	return blocks
}
