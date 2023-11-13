package data

import (
	"context"

	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/node/service/chain"
	"github.com/yuriykis/microblocknet/node/store"
)

type Retriever interface {
	GetBlockByHeight(ctx context.Context, height int) (*proto.Block, error)
	GetUTXOsByAddress(ctx context.Context, address []byte) ([]*proto.UTXO, error)
	Mempool() *Mempool
	MempoolList() []*proto.Transaction
	Chain() *chain.Chain
}

type dataRetriever struct {
	mempool *Mempool
	chain   *chain.Chain
}

func NewRetriever() Retriever {
	var s store.Storer = store.NewChainMemoryStore()

	chain := chain.New(s)

	return &dataRetriever{
		mempool: NewMempool(),
		chain:   chain,
	}
}

func (r *dataRetriever) GetBlockByHeight(ctx context.Context, height int) (*proto.Block, error) {
	return r.chain.GetBlockByHeight(height)
}

func (r *dataRetriever) GetUTXOsByAddress(ctx context.Context, address []byte) ([]*proto.UTXO, error) {
	return r.chain.Store().UTXOStore().GetByAddress(address)
}

func (r *dataRetriever) Mempool() *Mempool {
	return r.mempool
}

func (r *dataRetriever) MempoolList() []*proto.Transaction {
	return r.mempool.list()
}

func (r *dataRetriever) Chain() *chain.Chain {
	return r.chain
}
