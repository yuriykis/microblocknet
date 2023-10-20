package node

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuriykis/microblocknet/crypto"
	"github.com/yuriykis/microblocknet/proto"
	"github.com/yuriykis/microblocknet/store"
	"github.com/yuriykis/microblocknet/types"
	"github.com/yuriykis/microblocknet/util"
)

func TestNewChain(t *testing.T) {
	var (
		txStore    = store.NewMemoryTxStore()
		blockStore = store.NewMemoryBlockStore()
		utxoStore  = store.NewMemoryUTXOStore()
	)

	chain := NewChain(txStore, blockStore, utxoStore)
	assert.Equal(t, 0, chain.Height())

	assert.Equal(t, 1, len(chain.headers.headers))
	_, err := chain.GetBlockByHeight(0)
	assert.NoError(t, err)
}

func TestChainAddBlock(t *testing.T) {
	chain := NewChain(store.NewMemoryTxStore(), store.NewMemoryBlockStore(), store.NewMemoryUTXOStore())
	assert.Equal(t, 0, chain.Height())

	for i := 0; i < 10; i++ {
		block := util.RandomBlock()
		chain.addBlock(block)
		hashBlock := types.HashBlock(block)
		blockFromStore, err := chain.GetBlockByHash(hashBlock)
		assert.NoError(t, err)
		assert.Equal(t, block, blockFromStore)
		assert.Equal(t, i+1, chain.Height())
	}
}

func TestChainAddBlockWithTxs(t *testing.T) {
	chain := NewChain(store.NewMemoryTxStore(), store.NewMemoryBlockStore(), store.NewMemoryUTXOStore())
	assert.Equal(t, 0, chain.Height())
	myPrivKey := crypto.GeneratePrivateKey()
	toAddress := crypto.GeneratePrivateKey().PublicKey().Address()

	currentValue := int64(100000)
	// for i := 0; i < 10; i++ {
	prevBlock, err := chain.GetBlockByHeight(0)
	assert.NoError(t, err)
	prevBlockTx := prevBlock.GetTransactions()[len(prevBlock.GetTransactions())-1]
	assert.NotNil(t, prevBlockTx)

	block := util.RandomBlock()
	inputs := []*proto.TxInput{
		{
			PublicKey:  myPrivKey.PublicKey().Bytes(),
			PrevTxHash: []byte(types.HashTransaction(prevBlockTx)),
			OutIndex:   0,
		},
	}
	// fix the outputs not being recognized as an inputs in the next transaction
	outputs := []*proto.TxOutput{
		{
			Value:   100,
			Address: toAddress.Bytes(),
		},
		{
			Value:   currentValue - 100,
			Address: myPrivKey.PublicKey().Address().Bytes(),
		},
	}
	tx := &proto.Transaction{
		Inputs:  inputs,
		Outputs: outputs,
	}
	sig := types.SignTransaction(tx, myPrivKey)
	tx.Inputs[0].Signature = sig.Bytes()

	block.Transactions = append(block.Transactions, tx)
	block.Header.PrevBlockHash = []byte(types.HashBlock(prevBlock))
	block.Header.Height = int32(1)

	types.SignBlock(block, myPrivKey)

	err = chain.AddBlock(block)
	assert.NoError(t, err)
	// }
}
