package client

import (
	"context"

	"github.com/yuriykis/microblocknet/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	Endpoint string
	client   proto.NodeClient
}

func NewGRPCClient(endpoint string) (*GRPCClient, error) {
	conn, err := grpc.Dial(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	c := proto.NewNodeClient(conn)
	return &GRPCClient{
		Endpoint: endpoint,
		client:   c,
	}, nil
}

func (c *GRPCClient) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	return c.client.Handshake(ctx, v)
}

func (c *GRPCClient) NewTransaction(ctx context.Context, t *proto.Transaction) (*proto.Transaction, error) {
	return c.client.NewTransaction(ctx, t)
}

func (c *GRPCClient) NewBlock(ctx context.Context, b *proto.Block) (*proto.Block, error) {
	return c.client.NewBlock(ctx, b)
}
