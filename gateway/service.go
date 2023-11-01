package main

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/common/requests"
	"github.com/yuriykis/microblocknet/node/secure"
	nodeapi "github.com/yuriykis/microblocknet/node/service/api_client"
)

type service struct {
	nodeapi nodeapi.Client
}

func newService() *service {
	return &service{
		nodeapi: nodeapi.NewHTTPClient("http://localhost:4000"),
	}
}

func (s *service) BlockByHeight(ctx context.Context, height int) (*proto.Block, error) {
	b, err := s.nodeapi.GetBlockByHeight(ctx, height)
	if err != nil {
		return nil, err
	}
	return b.Block, nil
}

func (s *service) UTXOsByAddress(ctx context.Context, address []byte) ([]*proto.UTXO, error) {
	utxos, err := s.nodeapi.GetUTXOsByAddress(ctx, address)
	if err != nil {
		return nil, err
	}
	return utxos.UTXOs, nil
}

func (s *service) InitTransaction(
	ctx context.Context,
	t *Transaction,
) (*proto.Transaction, error) {
	clientUTXOs, err := s.nodeapi.GetUTXOsByAddress(ctx, t.FromAddress)
	if err != nil {
		return nil, err
	}
	if clientUTXOs == nil {
		return nil, err
	}
	var totalAmount int
	for _, utxo := range clientUTXOs.UTXOs {
		totalAmount += int(utxo.Output.Value)
	}
	if totalAmount < t.Amount {
		return nil, err
	}
	prevBlockRes, err := s.nodeapi.GetBlockByHeight(ctx, 0)
	if err != nil {
		log.Fatal(err)
	}
	prevBlock := prevBlockRes.Block
	prevBlockTx := prevBlock.GetTransactions()[len(prevBlock.GetTransactions())-1]
	txInput := &proto.TxInput{
		PrevTxHash: []byte(secure.HashTransaction(prevBlockTx)),
		PublicKey:  t.FromPubKey,
		OutIndex:   clientUTXOs.UTXOs[0].OutIndex,
	}
	txOutput1 := &proto.TxOutput{
		Value:   int64(t.Amount),
		Address: t.ToAddress,
	}
	txOutput2 := &proto.TxOutput{
		Value:   int64(totalAmount - t.Amount),
		Address: t.FromAddress,
	}
	return &proto.Transaction{
		Inputs:  []*proto.TxInput{txInput},
		Outputs: []*proto.TxOutput{txOutput1, txOutput2},
	}, nil
}

func (s *service) NewTransaction(
	ctx context.Context,
	t *proto.Transaction,
) (*proto.Transaction, error) {
	req := requests.NewTransactionRequest{
		Transaction: t,
	}
	res, err := s.nodeapi.NewTransaction(ctx, req)
	if err != nil {
		return nil, err
	}
	return res.Transaction, nil
}
