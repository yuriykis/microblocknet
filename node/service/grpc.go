package service

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/yuriykis/microblocknet/common/proto"
	"google.golang.org/grpc"
)

type GRPCMetricsHandler struct {
	reqCounter prometheus.Counter
	reqLatency prometheus.Histogram
	reqError   prometheus.Counter
}

func newGRPCMetricsHandler(reqName string) *GRPCMetricsHandler {
	reqCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_count", reqName),
		Help: fmt.Sprintf("Number of %s", reqName),
	})
	reqLatency := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    fmt.Sprintf("%s_latency", reqName),
		Help:    fmt.Sprintf("Latency of %s", reqName),
		Buckets: prometheus.LinearBuckets(0, 1, 10),
	})
	reqError := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_error", reqName),
		Help: fmt.Sprintf("Number of %s errors", reqName),
	})
	return &GRPCMetricsHandler{
		reqCounter: reqCounter,
		reqLatency: reqLatency,
		reqError:   reqError,
	}
}

func (h *GRPCMetricsHandler) instrument(ctx context.Context, reqName string, next func() error) error {
	h.reqCounter.Inc()
	start := time.Now()
	err := next()
	h.reqLatency.Observe(time.Since(start).Seconds())
	if err != nil {
		h.reqError.Inc()
	}
	return err
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

func startGRPCTransport(n NodeServer) error {
	s, ok := n.(*GRPCNodeServer)
	if !ok {
		return fmt.Errorf("invalid GRPCNodeServer")
	}
	ln, err := net.Listen("tcp", s.nodeListenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()
	return s.grpcServer.Serve(ln)
}

func stopGRPCTransport(n NodeServer) error {
	s, ok := n.(*GRPCNodeServer)
	if !ok {
		return fmt.Errorf("invalid GRPCNodeServer")
	}
	s.grpcServer.GracefulStop()
	return nil
}
