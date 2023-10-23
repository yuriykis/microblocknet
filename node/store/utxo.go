package store

import (
	"bytes"
	"sync"

	"github.com/yuriykis/microblocknet/node/proto"
	"github.com/yuriykis/microblocknet/node/types"
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

func (m *MemoryUTXOStore) List() []*proto.UTXO {
	m.lock.RLock()
	defer m.lock.RUnlock()

	utxos := make([]*proto.UTXO, 0)

	for _, utxo := range m.utxos {
		utxos = append(utxos, utxo)
	}

	return utxos
}

func (m *MemoryUTXOStore) GetByAddress(address []byte) ([]*proto.UTXO, error) {
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
