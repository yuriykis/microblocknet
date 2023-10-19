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
