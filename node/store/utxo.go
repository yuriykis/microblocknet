package store

import (
	"bytes"
	"context"
	"sync"

	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/node/secure"
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

func (m *MongoUTXOStore) Put(ctx context.Context, utxo *proto.UTXO) error {
	// TODO: implement
	return nil
}

func (m *MongoUTXOStore) Get(ctx context.Context, key string) (*proto.UTXO, error) {
	// TODO: implement
	return nil, nil
}

func (m *MongoUTXOStore) List(ctx context.Context) []*proto.UTXO {
	// TODO: implement
	return nil
}

func (m *MongoUTXOStore) GetByAddress(ctx context.Context, address []byte) ([]*proto.UTXO, error) {
	// TODO: implement
	return nil, nil
}
