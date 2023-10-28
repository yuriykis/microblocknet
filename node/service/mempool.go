package service

import (
	"sync"

	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/node/secure"
)

type Mempool struct {
	lock sync.RWMutex
	txs  map[string]*proto.Transaction
}

func NewMempool() *Mempool {
	return &Mempool{
		txs: make(map[string]*proto.Transaction),
	}
}

func (m *Mempool) Add(tx *proto.Transaction) {
	m.lock.Lock()
	defer m.lock.Unlock()
	hashTx := string(secure.HashTransaction(tx))
	m.txs[hashTx] = tx
}

func (m *Mempool) Contains(tx *proto.Transaction) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	hashTx := string(secure.HashTransaction(tx))
	_, ok := m.txs[hashTx]
	return ok
}

func (m *Mempool) Clear() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.txs = make(map[string]*proto.Transaction)
}

func (m *Mempool) Remove(tx *proto.Transaction) {
	m.lock.Lock()
	defer m.lock.Unlock()
	hashTx := string(secure.HashTransaction(tx))
	delete(m.txs, hashTx)
}

func (m *Mempool) list() []*proto.Transaction {
	m.lock.RLock()
	defer m.lock.RUnlock()
	txs := make([]*proto.Transaction, 0)
	for _, tx := range m.txs {
		txs = append(txs, tx)
	}
	return txs
}
