package node

import (
	"context"
	"log"
	"net"

	"github.com/yuriykis/microblocknet/proto"
	"google.golang.org/grpc"
)

type Node struct {
	ListenAddress string

	proto.UnimplementedNodeServer
}

func NewNode(listenAddress string) *Node {
	return &Node{
		ListenAddress: listenAddress,
	}
}

func (n *Node) Start() error {
	var (
		opt        = []grpc.ServerOption{}
		grpcServer = grpc.NewServer(opt...)
	)
	ln, err := net.Listen("tcp", n.ListenAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	proto.RegisterNodeServer(grpcServer, n)
	return grpcServer.Serve(ln)
}

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	return &proto.Version{
		Version: "0.0.1",
	}, nil
}

func (n *Node) NewTransaction(
	ctx context.Context,
	t *proto.Transaction,
) (*proto.Transaction, error) {
	return &proto.Transaction{}, nil
}

func (n *Node) NewBlock(ctx context.Context, b *proto.Block) (*proto.Block, error) {
	return &proto.Block{}, nil
}
