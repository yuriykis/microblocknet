package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuriykis/microblocknet/node/secure"
	"github.com/yuriykis/microblocknet/node/util"
)

func TestPutBlock(t *testing.T) {
	blockStore := NewMemoryBlockStore()
	putBlock := util.RandomBlock()
	err := blockStore.Put(putBlock)
	assert.Nil(t, err)

	hash := secure.HashBlock(putBlock)
	getBlock, err := blockStore.Get(hash)
	assert.Nil(t, err)
	assert.Equal(t, putBlock, getBlock)
}

func TestListBlocks(t *testing.T) {
	blockStore := NewMemoryBlockStore()
	firstBlock := util.RandomBlock()
	err := blockStore.Put(firstBlock)
	assert.Nil(t, err)

	secondBlock := util.RandomBlock()
	err = blockStore.Put(secondBlock)
	assert.Nil(t, err)

	blocks := blockStore.List()
	assert.Equal(t, 2, len(blocks))
	assert.Contains(t, blocks, firstBlock)
	assert.Contains(t, blocks, secondBlock)
}
