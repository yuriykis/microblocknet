package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuriykis/microblocknet/node/secure"
	"github.com/yuriykis/microblocknet/node/util"
)

func TestPutTransaction(t *testing.T) {
	txStore := NewMemoryTxStore()
	putTx := util.RandomTransaction()
	err := txStore.Put(putTx)
	assert.Nil(t, err)

	hash := secure.HashTransaction(putTx)
	getTx, err := txStore.Get(hash)
	assert.Nil(t, err)
	assert.Equal(t, putTx, getTx)
}

func TestListTransactions(t *testing.T) {
	txStore := NewMemoryTxStore()
	firstTx := util.RandomTransaction()
	err := txStore.Put(firstTx)
	assert.Nil(t, err)

	secondTx := util.RandomTransaction()
	err = txStore.Put(secondTx)
	assert.Nil(t, err)

	txs := txStore.List()
	assert.Equal(t, 2, len(txs))
	assert.Contains(t, txs, firstTx)
	assert.Contains(t, txs, secondTx)
}
