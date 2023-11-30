package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	// TODO: implement
}
