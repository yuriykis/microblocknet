package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuriykis/microblocknet/node/secure"
	"github.com/yuriykis/microblocknet/node/util"
)

func TestPutBlock(t *testing.T) {
	ctx := context.Background()
	blockStore := NewMemoryBlockStore()
	putBlock := util.RandomBlock()
	err := blockStore.Put(ctx, putBlock)
	assert.Nil(t, err)

	hash := secure.HashBlock(putBlock)
	getBlock, err := blockStore.Get(ctx, hash)
	assert.Nil(t, err)
	assert.Equal(t, putBlock, getBlock)
}

func TestListBlocks(t *testing.T) {
	ctx := context.Background()
	blockStore := NewMemoryBlockStore()
	firstBlock := util.RandomBlock()
	err := blockStore.Put(ctx, firstBlock)
	assert.Nil(t, err)

	secondBlock := util.RandomBlock()
	err = blockStore.Put(ctx, secondBlock)
	assert.Nil(t, err)

	blocks := blockStore.List(ctx)
	assert.Equal(t, 2, len(blocks))
	assert.Contains(t, blocks, firstBlock)
	assert.Contains(t, blocks, secondBlock)
}
