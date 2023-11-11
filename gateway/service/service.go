package service

import (
	"context"

	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/common/requests"
	"github.com/yuriykis/microblocknet/gateway/network"
	"github.com/yuriykis/microblocknet/gateway/types"
	"go.uber.org/zap"
)

type Service interface {
	BlockByHeight(ctx context.Context, height int) (*proto.Block, error)
	UTXOsByAddress(ctx context.Context, address []byte) ([]*proto.UTXO, error)
	InitTransaction(ctx context.Context, t *types.Transaction) (*proto.Transaction, error)
	NewTransaction(ctx context.Context, t *proto.Transaction) (*proto.Transaction, error)
}

type service struct {
	n      network.Networker
	logger *zap.SugaredLogger
}

func New(logger *zap.SugaredLogger, n network.Networker) Service {
	s := &service{
		n:      n,
		logger: logger,
	}
	return s
}

func (s *service) BlockByHeight(ctx context.Context, height int) (*proto.Block, error) {
	n, err := s.n.Node()
	if err != nil {
		return nil, err
	}
	b, err := n.GetBlockByHeight(ctx, height)
	if err != nil {
		return nil, err
	}
	return b.Block, nil
}

func (s *service) UTXOsByAddress(ctx context.Context, address []byte) ([]*proto.UTXO, error) {
	n, err := s.n.Node()
	if err != nil {
		return nil, err
	}
	utxos, err := n.GetUTXOsByAddress(ctx, address)
	if err != nil {
		return nil, err
	}
	return utxos.UTXOs, nil
}

func (s *service) InitTransaction(
	ctx context.Context,
	t *types.Transaction,
) (*proto.Transaction, error) {
	n, err := s.n.Node()
	if err != nil {
		return nil, err
	}
	clientUTXOs, err := n.GetUTXOsByAddress(ctx, t.FromAddress)
	if err != nil {
		return nil, err
	}
	if clientUTXOs == nil {
		return nil, err
	}
	heightRes, err := n.Height(ctx)
	if err != nil {
		return nil, err
	}
	prevBlockRes, err := n.GetBlockByHeight(ctx, heightRes.Height)
	if err != nil {
		s.logger.Errorf("failed to get block by height: %v", err)
		return nil, err
	}
	prevBlock := prevBlockRes.Block

	txBuilder := types.NewTransactionBuilder().
		SetClientUTXOs(clientUTXOs.UTXOs).
		SetChainHeight(heightRes.Height).
		SetPrevBlock(prevBlock).
		SetTransaction(t)
	tx, err := txBuilder.Build()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (s *service) NewTransaction(
	ctx context.Context,
	t *proto.Transaction,
) (*proto.Transaction, error) {
	req := requests.NewTransactionRequest{
		Transaction: t,
	}
	n, err := s.n.Node()
	if err != nil {
		return nil, err
	}
	res, err := n.NewTransaction(ctx, req)
	if err != nil {
		s.logger.Errorf("failed to send transaction: %v", err)
		return nil, err
	}
	return res.Transaction, nil
}
