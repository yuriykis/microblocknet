package node

import (
	"sync"

	"github.com/yuriykis/microblocknet/proto"
	"github.com/yuriykis/microblocknet/types"
)

type Mempool struct {
	lock sync.RWMutex
	txx  map[string]*proto.Transaction
}

func NewMempool() *Mempool {
	return &Mempool{
		txx: make(map[string]*proto.Transaction),
	}
}

func (m *Mempool) Add(tx *proto.Transaction) {
	m.lock.Lock()
	defer m.lock.Unlock()
	hashTx := string(types.HashTransaction(tx))
	m.txx[hashTx] = tx
}

func (m *Mempool) Contains(tx *proto.Transaction) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	hashTx := string(types.HashTransaction(tx))
	_, ok := m.txx[hashTx]
	return ok
}

func (m *Mempool) list() []*proto.Transaction {
	m.lock.RLock()
	defer m.lock.RUnlock()
	txx := make([]*proto.Transaction, 0)
	for _, tx := range m.txx {
		txx = append(txx, tx)
	}
	return txx
}