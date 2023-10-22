package node

import (
	"fmt"
	"log"
	"net"

	"github.com/yuriykis/microblocknet/proto"
	"google.golang.org/grpc"
)

// the Server interface is used to abstract the transport layer
type Server interface {
	Serve() error
	Close() error
}

type GRPCServer struct {
	grpcServer *grpc.Server
	listenAddr string

	chainServer ChainServer
}

func NewGRPCServer(grpcServer *grpc.Server, listenAddr string) *GRPCServer {
	return &GRPCServer{
		grpcServer: grpcServer,
		listenAddr: listenAddr,
	}
}

func (s *GRPCServer) Serve() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()
	return s.grpcServer.Serve(ln)
}

func (s *GRPCServer) Close() error {
	s.grpcServer.GracefulStop()
	return nil
}

func (s *GRPCServer) MakeTransport(listenAddr string, svc Node) *GRPCNodeServer {
	fmt.Printf("Node %s, starting GRPC transport\n", listenAddr)
	var (
		opt        = []grpc.ServerOption{}
		grpcServer = grpc.NewServer(opt...)
	)

	grpcNodeServer := NewGRPCNodeServer(svc)
	proto.RegisterNodeServer(grpcServer, grpcNodeServer)
	s.chainServer = grpcNodeServer

	return grpcNodeServer
}
