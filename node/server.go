package node

import (
	"context"

	"github.com/yuriykis/microblocknet/proto"
)

type Server interface {
	Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error)
	NewTransaction(ctx context.Context, t *proto.Transaction) (*proto.Transaction, error)
	NewBlock(ctx context.Context, b *proto.Block) (*proto.Block, error)
	Serve() error
	Close() error
}