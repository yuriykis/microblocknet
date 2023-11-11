package types

import (
	"fmt"

	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/node/secure"
)

type TransactionBuilder struct {
	clientUTXOs []*proto.UTXO
	chainHeight int
	prevBlock   *proto.Block
	t           *Transaction
}

func NewTransactionBuilder() *TransactionBuilder {
	return &TransactionBuilder{
		clientUTXOs: make([]*proto.UTXO, 0),
	}
}

func (tb *TransactionBuilder) SetClientUTXOs(utxos []*proto.UTXO) *TransactionBuilder {
	tb.clientUTXOs = utxos
	return tb
}

func (tb *TransactionBuilder) SetChainHeight(height int) *TransactionBuilder {
	tb.chainHeight = height
	return tb
}

func (tb *TransactionBuilder) SetPrevBlock(block *proto.Block) *TransactionBuilder {
	tb.prevBlock = block
	return tb
}

func (tb *TransactionBuilder) SetTransaction(t *Transaction) *TransactionBuilder {
	tb.t = t
	return tb
}

func (tb *TransactionBuilder) Build() (*proto.Transaction, error) {
	var totalAmount int
	for _, utxo := range tb.clientUTXOs {
		totalAmount += int(utxo.Output.Value)
	}
	if totalAmount < tb.t.Amount {
		return nil, fmt.Errorf("not enough funds")
	}
	prevBlockTx := tb.prevBlock.GetTransactions()[len(tb.prevBlock.GetTransactions())-1]
	txInput := &proto.TxInput{
		PrevTxHash: []byte(secure.HashTransaction(prevBlockTx)),
		PublicKey:  tb.t.FromPubKey,
		OutIndex:   tb.clientUTXOs[0].OutIndex,
	}
	txOutput1 := &proto.TxOutput{
		Value:   int64(tb.t.Amount),
		Address: tb.t.ToAddress,
	}
	txOutput2 := &proto.TxOutput{
		Value:   int64(totalAmount - tb.t.Amount),
		Address: tb.t.FromAddress,
	}
	return &proto.Transaction{
		Inputs:  []*proto.TxInput{txInput},
		Outputs: []*proto.TxOutput{txOutput1, txOutput2},
	}, nil
}
