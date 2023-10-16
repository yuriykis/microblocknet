package node

import (
	"context"

	"github.com/yuriykis/microblocknet/proto"
)

type GRPCNodeServer struct {
	proto.UnimplementedNodeServer
	svc Node
}

func NewGRPCNodeServer(svc Node) *GRPCNodeServer {
	return &GRPCNodeServer{
		svc: svc,
	}
}

func (s *GRPCNodeServer) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	return s.svc.Handshake(ctx, v)
}

func (s *GRPCNodeServer) NewTransaction(ctx context.Context, t *proto.Transaction) (*proto.Transaction, error) {
	return s.svc.NewTransaction(ctx, t)
}

func (s *GRPCNodeServer) NewBlock(ctx context.Context, b *proto.Block) (*proto.Block, error) {
	return s.svc.NewBlock(ctx, b)
}
