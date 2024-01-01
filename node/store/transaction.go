package store

import (
	"fmt"
	"sync"

	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/node/secure"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	txColl = "transaction"
)

// MemoryTxStore
type MemoryTxStore struct {
	lock sync.RWMutex
	txs  map[string]*proto.Transaction
}

func NewMemoryTxStore() *MemoryTxStore {
	return &MemoryTxStore{
		txs: make(map[string]*proto.Transaction),
	}
}

func (m *MemoryTxStore) Put(tx *proto.Transaction) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	hashTx := secure.HashTransaction(tx)
	m.txs[hashTx] = tx
	return nil
}

func (m *MemoryTxStore) Get(txHash string) (*proto.Transaction, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	tx, ok := m.txs[txHash]
	if !ok {
		return nil, fmt.Errorf("transaction with id %s not found", txHash)
	}
	return tx, nil
}

func (m *MemoryTxStore) List() []*proto.Transaction {
	m.lock.RLock()
	defer m.lock.RUnlock()
	txs := make([]*proto.Transaction, 0)
	for _, tx := range m.txs {
		txs = append(txs, tx)
	}
	return txs
}

// -----------------------------------------------------------------------------
// MongoTxStore

type MongoTxStore struct {
	client *mongo.Client
	coll   *mongo.Collection
}

func NewMongoTxStore(client *mongo.Client) *MongoTxStore {
	return &MongoTxStore{
		client: client,
		coll:   client.Database(mongoDBName).Collection(txColl),
	}
}

func (m *MongoTxStore) Put(tx *proto.Transaction) error {
	// TODO: implement
	return nil
}

func (m *MongoTxStore) Get(txID string) (*proto.Transaction, error) {
	// TODO: implement
	return nil, nil
}

func (m *MongoTxStore) List() []*proto.Transaction {
	// TODO: implement
	return nil
}
