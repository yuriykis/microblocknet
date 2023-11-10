package main

import (
	"context"
	"fmt"
	"time"

	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/common/requests"
	"github.com/yuriykis/microblocknet/node/secure"
	"go.uber.org/zap"
)

const pingPeersInterval = 5 * time.Second

type service struct {
	*nodeapi
	logger *zap.SugaredLogger
}

func newService(logger *zap.SugaredLogger) *service {
	s := &service{
		nodeapi: newNodeAPI(logger),
		logger:  logger,
	}
	go s.pingKnownPeers()
	return s
}

func (s *service) BlockByHeight(ctx context.Context, height int) (*proto.Block, error) {
	if s.nodeApi() == nil {
		return nil, fmt.Errorf("node api is unavailable")
	}
	b, err := s.nodeApi().GetBlockByHeight(ctx, height)
	if err != nil {
		return nil, err
	}
	return b.Block, nil
}

func (s *service) UTXOsByAddress(ctx context.Context, address []byte) ([]*proto.UTXO, error) {
	if s.nodeApi() == nil {
		return nil, fmt.Errorf("node api is unavailable")
	}
	utxos, err := s.nodeApi().GetUTXOsByAddress(ctx, address)
	if err != nil {
		return nil, err
	}
	return utxos.UTXOs, nil
}

func (s *service) InitTransaction(
	ctx context.Context,
	t *Transaction,
) (*proto.Transaction, error) {
	if s.nodeApi() == nil {
		return nil, fmt.Errorf("node api is unavailable")
	}
	clientUTXOs, err := s.nodeApi().GetUTXOsByAddress(ctx, t.FromAddress)
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
	heightRes, err := s.nodeApi().Height(ctx)
	if err != nil {
		return nil, err
	}
	prevBlockRes, err := s.nodeApi().GetBlockByHeight(ctx, heightRes.Height)
	if err != nil {
		s.logger.Errorf("failed to get block by height: %v", err)
		return nil, err
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
	if s.nodeApi() == nil {
		return nil, fmt.Errorf("node api is unavailable")
	}
	res, err := s.nodeApi().NewTransaction(ctx, req)
	if err != nil {
		s.logger.Errorf("failed to send transaction: %v", err)
		return nil, err
	}
	return res.Transaction, nil
}

func (s *service) NewNode(addr string) error {
	s.NewHost(addr)
	return nil
}

func (s *service) pingKnownPeers() {
	s.logger.Infof("starting to ping known peers")
	for {
		select {
		case <-time.After(pingPeersInterval):
			for _, addr := range s.Peers() {
				if err := s.pingPeer(addr); err != nil {
					s.logger.Errorf("failed to ping peer: %v", err)
				}
			}
		}
	}
}
