package service

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/yuriykis/microblocknet/node/proto"
	"google.golang.org/grpc"
)

// used to manage the transport layer between nodes
type TransportServer interface {
	Start() error
	Stop() error
}

// interface for the node business logic
type NodeServer interface {
	Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error)
	NewTransaction(ctx context.Context, t *proto.Transaction) (*proto.Transaction, error)
	NewBlock(ctx context.Context, b *proto.Block) (*proto.Block, error)
	GetBlocks(ctx context.Context, v *proto.Version) (*proto.Blocks, error)
}

// GRPCNodeServer implements the NodeServer interface and the TransportServer interface
type GRPCNodeServer struct {
	proto.UnimplementedNodeServer
	node Node

	grpcServer     *grpc.Server
	nodeListenAddr string
}

func NewGRPCNodeServer(node Node, nodeListenAddr string) *GRPCNodeServer {
	fmt.Printf("Node %s, starting GRPC transport\n", nodeListenAddr)
	var (
		opt        = []grpc.ServerOption{}
		grpcServer = grpc.NewServer(opt...)
	)
	grpcNodeServer := &GRPCNodeServer{
		node:           node,
		grpcServer:     grpcServer,
		nodeListenAddr: nodeListenAddr,
	}
	grpcNodeServer.grpcServer = grpcServer
	proto.RegisterNodeServer(grpcNodeServer.grpcServer, grpcNodeServer)
	return grpcNodeServer
}

// service methods (NodeServer interface)
func (s *GRPCNodeServer) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	return s.node.Handshake(ctx, v)
}

func (s *GRPCNodeServer) NewTransaction(ctx context.Context, t *proto.Transaction) (*proto.Transaction, error) {
	return s.node.NewTransaction(ctx, t)
}

func (s *GRPCNodeServer) NewBlock(ctx context.Context, b *proto.Block) (*proto.Block, error) {
	return s.node.NewBlock(ctx, b)
}

func (s *GRPCNodeServer) GetBlocks(ctx context.Context, v *proto.Version) (*proto.Blocks, error) {
	return s.node.GetBlocks(ctx, v)
}

// transport methods (TransportServer interface)
func (s *GRPCNodeServer) Start() error {
	ln, err := net.Listen("tcp", s.nodeListenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()
	return s.grpcServer.Serve(ln)
}

func (s *GRPCNodeServer) Stop() error {
	s.grpcServer.GracefulStop()
	return nil
}