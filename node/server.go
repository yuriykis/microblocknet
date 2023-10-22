package node

import (
	"context"

	"github.com/yuriykis/microblocknet/proto"
)

// TODO: Create a separate interface for the server
type Server interface {
	Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error)
	NewTransaction(ctx context.Context, t *proto.Transaction) (*proto.Transaction, error)
	NewBlock(ctx context.Context, b *proto.Block) (*proto.Block, error)
	GetBlocks(ctx context.Context, v *proto.Version) (*proto.Blocks, error)
	Serve() error
	Close() error
}
