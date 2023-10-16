package node

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/yuriykis/microblocknet/proto"
	"google.golang.org/grpc"
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

func makeGRPCTransport(listenAddr string, svc Node) error {
	fmt.Printf("Node %s, starting GRPC transport\n", listenAddr)
	var (
		opt        = []grpc.ServerOption{}
		grpcServer = grpc.NewServer(opt...)
	)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()

	grpcNodeServer := NewGRPCNodeServer(svc)
	proto.RegisterNodeServer(grpcServer, grpcNodeServer)

	return grpcServer.Serve(ln)
}
