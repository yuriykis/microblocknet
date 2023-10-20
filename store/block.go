package store

import (
	"fmt"
	"sync"

	"github.com/yuriykis/microblocknet/proto"
	"github.com/yuriykis/microblocknet/types"
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

func (m *MemoryBlockStore) Put(block *proto.Block) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	hash := types.HashBlock(block)
	m.blocks[hash] = block
	return nil
}

func (m *MemoryBlockStore) Get(blockHash string) (*proto.Block, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	block, ok := m.blocks[blockHash]
	if !ok {
		return nil, fmt.Errorf("block with id %s not found", blockHash)
	}
	return block, nil
}
