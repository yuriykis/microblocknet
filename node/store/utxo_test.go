package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuriykis/microblocknet/common/crypto"
	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/node/secure"
	"github.com/yuriykis/microblocknet/node/util"
)

func TestPutUTXO(t *testing.T) {
	utxoStore := NewMemoryUTXOStore()
	putUTXO := util.RandomUTXO()
	err := utxoStore.Put(putUTXO)
	assert.Nil(t, err)

	utxoKey := secure.MakeUTXOKey(putUTXO.TxHash, int(putUTXO.OutIndex))
	getUTXO, err := utxoStore.Get(utxoKey)
	assert.Nil(t, err)
	assert.Equal(t, putUTXO, getUTXO)
}

func TestListUTXOs(t *testing.T) {
	utxoStore := NewMemoryUTXOStore()
	firstUTXO := util.RandomUTXO()
	err := utxoStore.Put(firstUTXO)
	assert.Nil(t, err)

	secondUTXO := util.RandomUTXO()
	err = utxoStore.Put(secondUTXO)
	assert.Nil(t, err)

	utxos := utxoStore.List()
	assert.Equal(t, 2, len(utxos))
	assert.Contains(t, utxos, firstUTXO)
	assert.Contains(t, utxos, secondUTXO)
}

func TestGetUTXOByAddress(t *testing.T) {
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
	err := txStore.Put(tx)
	assert.Nil(t, err)

	utxoStore := NewMemoryUTXOStore()
	utxo := &proto.UTXO{
		TxHash:   []byte(secure.HashTransaction(tx)),
		OutIndex: 0,
		Output:   tx.Outputs[0],
		Spent:    false,
	}
	err = utxoStore.Put(utxo)
	assert.Nil(t, err)

	utxos, err := utxoStore.GetByAddress(privKey.PublicKey().Address().Bytes())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(utxos))
	assert.Equal(t, utxo, utxos[0])

}
