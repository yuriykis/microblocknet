package node

import (
	"context"
	"log"
	"net"
	"sync"

	"github.com/yuriykis/microblocknet/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Node struct {
	ListenAddress string

	peersLock sync.RWMutex
	peers     map[proto.NodeClient]*proto.Version
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

func (n *Node) Version() *proto.Version {
	return &proto.Version{
		Version: "0.0.1",
	}
}

func (n *Node) AddPeer(c proto.NodeClient, v *proto.Version) {
	n.peersLock.Lock()
	defer n.peersLock.Unlock()
	n.peers[c] = v
}

func (n *Node) dialRemote(address string) (*proto.NodeClient, *proto.Version, error) {
	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, err
	}

	client := proto.NewNodeClient(conn)
	version, err := client.Handshake(context.Background(), n.Version())
	if err != nil {
		return nil, nil, err
	}

	return &client, version, nil
}

func (n *Node) bootstrapNetwork(addrs []string) error {
	for _, addr := range addrs {
		client, version, err := n.dialRemote(addr)
		if err != nil {
			return err
		}
		n.AddPeer(*client, version)
	}
	return nil
}
