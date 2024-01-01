package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuriykis/microblocknet/common/crypto"
	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/node/secure"
	"github.com/yuriykis/microblocknet/node/util"
)

func TestPutUTXO(t *testing.T) {
	ctx := context.Background()
	utxoStore := NewMemoryUTXOStore()
	putUTXO := util.RandomUTXO()
	err := utxoStore.Put(ctx, putUTXO)
	assert.Nil(t, err)

	utxoKey := secure.MakeUTXOKey(putUTXO.TxHash, int(putUTXO.OutIndex))
	getUTXO, err := utxoStore.Get(ctx, utxoKey)
	assert.Nil(t, err)
	assert.Equal(t, putUTXO, getUTXO)
}

func TestListUTXOs(t *testing.T) {
	ctx := context.Background()
	utxoStore := NewMemoryUTXOStore()
	firstUTXO := util.RandomUTXO()
	err := utxoStore.Put(ctx, firstUTXO)
	assert.Nil(t, err)

	secondUTXO := util.RandomUTXO()
	err = utxoStore.Put(ctx, secondUTXO)
	assert.Nil(t, err)

	utxos := utxoStore.List(ctx)
	assert.Equal(t, 2, len(utxos))
	assert.Contains(t, utxos, firstUTXO)
	assert.Contains(t, utxos, secondUTXO)
}

func TestGetUTXOByAddress(t *testing.T) {
	ctx := context.Background()
	txStore := NewMemoryTxStore()
	privKey := crypto.GeneratePrivateKey()
	tx := &proto.Transaction{
		Inputs: []*proto.TxInput{},
		Outputs: []*proto.TxOutput{
			{
				Value:   100000,
				Address: privKey.PublicKey().Address().Bytes(),
			},
		},
	}
	err := txStore.Put(ctx, tx)
	assert.Nil(t, err)

	utxoStore := NewMemoryUTXOStore()
	utxo := &proto.UTXO{
		TxHash:   []byte(secure.HashTransaction(tx)),
		OutIndex: 0,
		Output:   tx.Outputs[0],
		Spent:    false,
	}
	err = utxoStore.Put(ctx, utxo)
	assert.Nil(t, err)

	utxos, err := utxoStore.GetByAddress(ctx, privKey.PublicKey().Address().Bytes())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(utxos))
	assert.Equal(t, utxo, utxos[0])

}
