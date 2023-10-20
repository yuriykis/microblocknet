package node

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuriykis/microblocknet/store"
)

func TestNewChain(t *testing.T) {
	txStore := store.NewMemoryTxStore()
	blockStore := store.NewMemoryBlockStore()

	chain := NewChain(txStore, blockStore)
	assert.Equal(t, 0, chain.Height())
	// TODO: check genesis block
}
