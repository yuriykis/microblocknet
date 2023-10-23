package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuriykis/microblocknet/node/crypto"
	"github.com/yuriykis/microblocknet/node/proto"
	"github.com/yuriykis/microblocknet/node/store"
	"github.com/yuriykis/microblocknet/node/types"
	"github.com/yuriykis/microblocknet/node/util"
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
	myPrivKey := crypto.PrivateKeyFromString(godSeed)
	toAddress := crypto.GeneratePrivateKey().PublicKey().Address()

	currentValue := int64(100000)
	for i := 0; i < 100; i++ {
		prevBlock, err := chain.GetBlockByHeight(i)
		assert.NotNil(t, prevBlock)
		assert.Nil(t, err)
		prevBlockTx := prevBlock.GetTransactions()[len(prevBlock.GetTransactions())-1]
		assert.NotNil(t, prevBlockTx)

		block := util.RandomBlock()
		myUTXOs, err := chain.utxoStore.GetByAddress(myPrivKey.PublicKey().Address().Bytes())
		assert.NotNil(t, myUTXOs)
		assert.Nil(t, err)

		inputs := []*proto.TxInput{
			{
				PublicKey:  myPrivKey.PublicKey().Bytes(),
				PrevTxHash: []byte(types.HashTransaction(prevBlockTx)),
				OutIndex:   myUTXOs[0].OutIndex,
			},
		}
		currentValue = currentValue - 100
		outputs := []*proto.TxOutput{
			{
				Value:   100,
				Address: toAddress.Bytes(),
			},
			{
				Value:   currentValue,
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
		block.Header.Height = int32(i + 1)

		types.SignBlock(block, myPrivKey)

		err = chain.AddBlock(block)
		assert.Nil(t, err)
	}
}
