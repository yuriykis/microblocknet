package store

import (
	"bytes"
	"context"
	"encoding/hex"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/node/secure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	utxoColl = "utxo"
)

type MemoryUTXOStore struct {
	lock  sync.RWMutex
	utxos map[string]*proto.UTXO
}

func NewMemoryUTXOStore() *MemoryUTXOStore {
	return &MemoryUTXOStore{
		utxos: make(map[string]*proto.UTXO),
	}
}

func (m *MemoryUTXOStore) Put(ctx context.Context, utxo *proto.UTXO) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	key := secure.MakeUTXOKey(utxo.TxHash, int(utxo.OutIndex))

	m.utxos[key] = utxo

	return nil
}

func (m *MemoryUTXOStore) Get(ctx context.Context, key string) (*proto.UTXO, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	utxo, ok := m.utxos[key]
	if !ok {
		return nil, nil
	}
	return utxo, nil
}

func (m *MemoryUTXOStore) List(ctx context.Context) []*proto.UTXO {
	m.lock.RLock()
	defer m.lock.RUnlock()

	utxos := make([]*proto.UTXO, 0)

	for _, utxo := range m.utxos {
		utxos = append(utxos, utxo)
	}

	return utxos
}

func (m *MemoryUTXOStore) GetByAddress(ctx context.Context, address []byte) ([]*proto.UTXO, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	utxos := make([]*proto.UTXO, 0)

	for _, utxo := range m.utxos {
		if bytes.Equal(utxo.Output.Address, address) && !utxo.Spent {
			utxos = append(utxos, utxo)
		}
	}

	return utxos, nil
}

// -----------------------------------------------------------------------------

type MongoUTXOStore struct {
	client *mongo.Client
	coll   *mongo.Collection
}

func NewMongoUTXOStore(client *mongo.Client) *MongoUTXOStore {
	return &MongoUTXOStore{
		client: client,
		coll:   client.Database(mongoDBName).Collection(utxoColl),
	}
}

// Put inserts a new UTXO into the store, implements UTXOStorer interface
func (m *MongoUTXOStore) Put(ctx context.Context, utxo *proto.UTXO) error {
	key := secure.MakeUTXOKey(utxo.TxHash, int(utxo.OutIndex))
	res, err := m.coll.InsertOne(ctx, bson.M{
		"key":  hex.EncodeToString([]byte(key)),
		"utxo": utxo,
	})
	if err != nil {
		return err
	}
	logrus.Infof("inserted utxo %s", res.InsertedID)
	return nil
}

// Get retrieves a UTXO from the store, implements UTXOStorer interface
func (m *MongoUTXOStore) Get(ctx context.Context, key string) (*proto.UTXO, error) {
	var utxoDoc struct {
		Key  string     `bson:"key"`
		UTXO proto.UTXO `bson:"utxo"`
	}
	if err := m.coll.FindOne(ctx, bson.M{
		"key": hex.EncodeToString([]byte(key)),
	}).Decode(&utxoDoc); err != nil {
		return nil, err
	}
	return &utxoDoc.UTXO, nil
}

// List retrieves all UTXOs from the store, implements UTXOStorer interface
func (m *MongoUTXOStore) List(ctx context.Context) []*proto.UTXO {
	utxosDocs := make(
		[]struct {
			Key  string     `bson:"key"`
			UTXO proto.UTXO `bson:"utxo"`
		}, 0)
	cur, err := m.coll.Find(ctx, nil)
	if err != nil {
		return nil
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var utxoDoc struct {
			Key  string     `bson:"key"`
			UTXO proto.UTXO `bson:"utxo"`
		}
		if err := cur.Decode(&utxoDoc); err != nil {
			logrus.Errorf("error decoding utxo: %s", err)
			continue
		}
		utxosDocs = append(utxosDocs, utxoDoc)
	}
	if err := cur.Err(); err != nil {
		logrus.Errorf("error iterating utxos: %s", err)
		return nil
	}
	utxos := make([]*proto.UTXO, 0)
	for _, utxoDoc := range utxosDocs {
		utxos = append(utxos, &utxoDoc.UTXO)
	}
	return utxos
}

// GetByAddress retrieves all UTXOs for a given address from the store, implements UTXOStorer interface
func (m *MongoUTXOStore) GetByAddress(ctx context.Context, address []byte) ([]*proto.UTXO, error) {
	utxosDocs := make(
		[]struct {
			Key  string     `bson:"key"`
			UTXO proto.UTXO `bson:"utxo"`
		}, 0)
	cur, err := m.coll.Find(ctx, bson.M{
		"utxo.output.address": address,
	})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var utxoDoc struct {
			Key  string     `bson:"key"`
			UTXO proto.UTXO `bson:"utxo"`
		}
		if err := cur.Decode(&utxoDoc); err != nil {
			logrus.Errorf("error decoding utxo: %s", err)
			continue
		}
		utxosDocs = append(utxosDocs, utxoDoc)
	}
	if err := cur.Err(); err != nil {
		logrus.Errorf("error iterating utxos: %s", err)
		return nil, err
	}
	utxos := make([]*proto.UTXO, 0)
	for _, utxoDoc := range utxosDocs {
		utxos = append(utxos, &utxoDoc.UTXO)
	}
	return utxos, nil
}
