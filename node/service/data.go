package service

import (
	"context"

	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/node/store"
)

type DataRetriever interface {
	GetBlockByHeight(ctx context.Context, height int) (*proto.Block, error)
	GetUTXOsByAddress(ctx context.Context, address []byte) ([]*proto.UTXO, error)
	Mempool() *Mempool
	Chain() *Chain
}

type dataRetriever struct {
	mempool *Mempool
	chain   *Chain
}

func NewDataRetriever() DataRetriever {
	var (
		txStore    = store.NewMemoryTxStore()
		blockStore = store.NewMemoryBlockStore()
		utxoStore  = store.NewMemoryUTXOStore()
	)
	chain := NewChain(txStore, blockStore, utxoStore)

	return &dataRetriever{
		mempool: NewMempool(),
		chain:   chain,
	}
}

func (r *dataRetriever) GetBlockByHeight(ctx context.Context, height int) (*proto.Block, error) {
	return r.chain.GetBlockByHeight(height)
}

func (r *dataRetriever) GetUTXOsByAddress(ctx context.Context, address []byte) ([]*proto.UTXO, error) {
	return r.chain.utxoStore.GetByAddress(address)
}

func (r *dataRetriever) Mempool() *Mempool {
	return r.mempool
}

func (r *dataRetriever) Chain() *Chain {
	return r.chain
}
