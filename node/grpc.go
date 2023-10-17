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
	svc        Node
	grpcServer *grpc.Server
	listenAddr string
}

func NewGRPCNodeServer(svc Node, grpcServer *grpc.Server, listenAddr string) *GRPCNodeServer {
	return &GRPCNodeServer{
		svc:        svc,
		grpcServer: grpcServer,
		listenAddr: listenAddr,
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

func (s *GRPCNodeServer) Serve() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()
	return s.grpcServer.Serve(ln)
}

func (s *GRPCNodeServer) Close() error {
	s.grpcServer.GracefulStop()
	return nil
}

func MakeGRPCTransport(listenAddr string, svc Node) *GRPCNodeServer {
	fmt.Printf("Node %s, starting GRPC transport\n", listenAddr)
	var (
		opt        = []grpc.ServerOption{}
		grpcServer = grpc.NewServer(opt...)
	)

	grpcNodeServer := NewGRPCNodeServer(svc, grpcServer, listenAddr)
	proto.RegisterNodeServer(grpcServer, grpcNodeServer)

	return grpcNodeServer
}
