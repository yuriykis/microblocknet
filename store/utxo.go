package store

import (
	"sync"

	"github.com/yuriykis/microblocknet/proto"
	"github.com/yuriykis/microblocknet/types"
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

func (m *MemoryUTXOStore) Put(utxo *proto.UTXO) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	key := types.MakeUTXOKey(utxo.TxHash, int(utxo.OutIndex))

	m.utxos[key] = utxo

	return nil
}

func (m *MemoryUTXOStore) Get(key string) (*proto.UTXO, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	utxo, ok := m.utxos[key]
	if !ok {
		return nil, nil
	}
	return utxo, nil
}
