package store

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/node/secure"
	"go.mongodb.org/mongo-driver/bson"
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

func (m *MemoryTxStore) Put(ctx context.Context, tx *proto.Transaction) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	hashTx := secure.HashTransaction(tx)
	m.txs[hashTx] = tx
	return nil
}

func (m *MemoryTxStore) Get(ctx context.Context, txHash string) (*proto.Transaction, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	tx, ok := m.txs[txHash]
	if !ok {
		return nil, fmt.Errorf("transaction with id %s not found", txHash)
	}
	return tx, nil
}

func (m *MemoryTxStore) List(ctx context.Context) []*proto.Transaction {
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

func (m *MongoTxStore) Put(ctx context.Context, tx *proto.Transaction) error {
	txHash := secure.HashTransaction(tx)
	res, err := m.coll.InsertOne(ctx, bson.M{
		"txHash": txHash,
		"tx":     tx,
	})
	if err != nil {
		return err
	}
	logrus.Debugf("inserted transaction %s", res.InsertedID)
	return nil
}

func (m *MongoTxStore) Get(ctx context.Context, txHash string) (*proto.Transaction, error) {
	var txDoc struct {
		TxHash string            `bson:"txHash"`
		Tx     proto.Transaction `bson:"tx"`
	}
	if err := m.coll.FindOne(ctx, bson.M{
		"txHash": txHash,
	}).Decode(&txDoc); err != nil {
		return nil, err
	}
	return &txDoc.Tx, nil
}

func (m *MongoTxStore) List(ctx context.Context) []*proto.Transaction {
	txsDocs := make(
		[]struct {
			TxHash string            `bson:"txHash"`
			Tx     proto.Transaction `bson:"tx"`
		}, 0)

	cur, err := m.coll.Find(ctx, nil)
	if err != nil {
		return nil
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var txDoc struct {
			TxHash string            `bson:"txHash"`
			Tx     proto.Transaction `bson:"tx"`
		}
		if err := cur.Decode(&txDoc); err != nil {
			return nil
		}
		txsDocs = append(txsDocs, txDoc)
	}
	txs := make([]*proto.Transaction, len(txsDocs))
	for i, txDoc := range txsDocs {
		txs[i] = &txDoc.Tx
	}
	return txs
}
