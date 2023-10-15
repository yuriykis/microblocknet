package node

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"github.com/yuriykis/microblocknet/proto"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Node struct {
	ListenAddress string

	logger    *zap.SugaredLogger
	peersLock sync.RWMutex
	peers     map[proto.NodeClient]*proto.Version
	proto.UnimplementedNodeServer
}

func NewNode(listenAddress string) *Node {
	return &Node{
		ListenAddress: listenAddress,
		peers:         make(map[proto.NodeClient]*proto.Version),
		logger:        makeLogger(),
	}
}

func (n *Node) String() string {
	return n.ListenAddress
}

func (n *Node) Start(bootstrapNodes []string) error {
	var (
		opt        = []grpc.ServerOption{}
		grpcServer = grpc.NewServer(opt...)
	)
	ln, err := net.Listen("tcp", n.ListenAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()

	if len(bootstrapNodes) > 0 {
		go func() {
			if err := n.bootstrapNetwork(bootstrapNodes); err != nil {
				log.Fatalf("Node: %s, failed to bootstrap network: %v", n, err)
			}
		}()
	}
	proto.RegisterNodeServer(grpcServer, n)
	n.logger.Infof("Node: %s, listening on %s", n, n.ListenAddress)
	return grpcServer.Serve(ln)
}

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	c, err := makeNodeClient(v.ListenAddress)
	if err != nil {
		return nil, err
	}
	n.addPeer(c, v)
	log.Printf("Node: %s, sending handshake to %s", n, v.ListenAddress)
	return n.Version(), nil
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

func (n *Node) Version() *proto.Version {
	return &proto.Version{
		Version:       "0.0.1",
		ListenAddress: n.ListenAddress,
		Peers:         n.Peers(),
	}
}

func (n *Node) addPeer(c proto.NodeClient, v *proto.Version) {
	n.peersLock.Lock()
	defer n.peersLock.Unlock()
	n.peers[c] = v

	if len(v.Peers) > 0 {
		go func() {
			if err := n.bootstrapNetwork(v.Peers); err != nil {
				log.Fatalf("Node: %s, failed to bootstrap network: %v", n, err)
			}
		}()
	}
}

func (n *Node) dialRemote(address string) (*proto.NodeClient, *proto.Version, error) {
	client, err := makeNodeClient(address)
	if err != nil {
		return nil, nil, err
	}
	version, err := client.Handshake(context.Background(), n.Version())
	n.logger.Infof("Node: %s, connected to %s", n, address)
	if err != nil {
		return nil, nil, err
	}

	return &client, version, nil
}

func (n *Node) bootstrapNetwork(addrs []string) error {
	for _, addr := range addrs {
		if !n.canConnectWith(addr) {
			continue
		}
		client, version, err := n.dialRemote(addr)
		if err != nil {
			return err
		}
		n.addPeer(*client, version)
	}
	return nil
}

func (n *Node) canConnectWith(addr string) bool {
	n.peersLock.RLock()
	defer n.peersLock.RUnlock()
	if addr == n.ListenAddress {
		return false
	}
	for _, peer := range n.Peers() {
		if peer == addr {
			return false
		}
	}
	return true
}

func (n *Node) Peers() []string {
	n.peersLock.RLock()
	defer n.peersLock.RUnlock()
	peersList := make([]string, 0)
	for _, v := range n.peers {
		peersList = append(peersList, v.ListenAddress)
	}
	return peersList
}

func makeNodeClient(address string) (proto.NodeClient, error) {
	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return proto.NewNodeClient(conn), nil
}

func makeLogger() *zap.SugaredLogger {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.TimeKey = "timestamp"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339Nano)
	logger, err := loggerConfig.Build()
	if err != nil {
		log.Fatal(err)
	}
	return logger.Sugar()
}
