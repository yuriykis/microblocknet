package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuriykis/microblocknet/crypto"
	"github.com/yuriykis/microblocknet/proto"
	"github.com/yuriykis/microblocknet/util"
)

func TestHashTransaction(t *testing.T) {
	fromPrivKey := crypto.GeneratePrivateKey()
	fromAddress := fromPrivKey.PublicKey().Address().String()

	toPrivKey := crypto.GeneratePrivateKey()
	toAddress := toPrivKey.PublicKey().Address().String()

	txInput := &proto.TxInput{
		PrevTxHash: util.RandomHash(),
		PublicKey:  fromPrivKey.PublicKey().String(),
	}
	txOutput1 := &proto.TxOutput{
		Value:   100,
		Address: toAddress,
	}
	txOutput2 := &proto.TxOutput{
		Value:   900,
		Address: fromAddress,
	}
	tx := &proto.Transaction{
		Inputs:  []*proto.TxInput{txInput},
		Outputs: []*proto.TxOutput{txOutput1, txOutput2},
	}

	sig := SignTransaction(tx, fromPrivKey)
	txInput.Signature = sig.String()
	assert.True(t, VerifyTransaction(tx))
}
