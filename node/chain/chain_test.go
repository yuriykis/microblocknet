package chain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuriykis/microblocknet/common/crypto"
	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/node/secure"
	"github.com/yuriykis/microblocknet/node/store"
	"github.com/yuriykis/microblocknet/node/util"
)

func TestNewChain(t *testing.T) {
	s := store.NewChainMemoryStore()
	chain := New(s)
	assert.Equal(t, 0, chain.Height())

	assert.Equal(t, 1, len(chain.headers.headers))
	_, err := chain.GetBlockByHeight(0)
	assert.NoError(t, err)
}

func TestChainAddBlock(t *testing.T) {
	s := store.NewChainMemoryStore()
	chain := New(s)
	assert.Equal(t, 0, chain.Height())

	for i := 0; i < 10; i++ {
		block := util.RandomBlock()
		chain.addBlock(block)
		hashBlock := secure.HashBlock(block)
		blockFromStore, err := chain.GetBlockByHash(hashBlock)
		assert.NoError(t, err)
		assert.Equal(t, block, blockFromStore)
		assert.Equal(t, i+1, chain.Height())
	}
}

func TestChainAddBlockWithTxs(t *testing.T) {
	ctx := context.Background()
	s := store.NewChainMemoryStore()
	chain := New(s)
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
		myUTXOs, err := chain.Store().UTXOStore(ctx).GetByAddress(ctx, myPrivKey.PublicKey().Address().Bytes())
		assert.NotNil(t, myUTXOs)
		assert.Nil(t, err)

		inputs := []*proto.TxInput{
			{
				PublicKey:  myPrivKey.PublicKey().Bytes(),
				PrevTxHash: []byte(secure.HashTransaction(prevBlockTx)),
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
		sig := secure.SignTransaction(tx, myPrivKey)
		tx.Inputs[0].Signature = sig.Bytes()

		block.Transactions = append(block.Transactions, tx)
		block.Header.PrevBlockHash = []byte(secure.HashBlock(prevBlock))
		block.Header.Height = int32(i + 1)

		secure.SignBlock(block, myPrivKey)

		err = chain.AddBlock(block)
		assert.Nil(t, err)
	}
}
