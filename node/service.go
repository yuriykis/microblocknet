package node

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/yuriykis/microblocknet/node/client"
	"github.com/yuriykis/microblocknet/proto"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Node interface {
	Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error)
	NewTransaction(ctx context.Context, t *proto.Transaction) (*proto.Transaction, error)
	NewBlock(ctx context.Context, b *proto.Block) (*proto.Block, error)
}

type peersMap struct {
	peers map[client.Client]*proto.Version
	lock  sync.RWMutex
}

func NewPeersMap() *peersMap {
	return &peersMap{
		peers: make(map[client.Client]*proto.Version),
	}
}
func (pm *peersMap) addPeer(c client.Client, v *proto.Version) {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	pm.peers[c] = v
}

func (pm *peersMap) List() []string {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	peersList := make([]string, 0)
	for _, v := range pm.peers {
		peersList = append(peersList, v.ListenAddress)
	}
	return peersList
}

type NetNode struct {
	ListenAddress string

	logger *zap.SugaredLogger
	peers  *peersMap
	proto.UnimplementedNodeServer
}

func New(listenAddress string) *NetNode {
	return &NetNode{
		ListenAddress: listenAddress,
		peers:         NewPeersMap(),
		logger:        makeLogger(),
	}
}

func (n *NetNode) String() string {
	return n.ListenAddress
}

func (n *NetNode) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	c, err := client.NewGRPCClient(v.ListenAddress)
	if err != nil {
		return nil, err
	}
	n.addPeer(c, v)
	log.Printf("NetNode: %s, sending handshake to %s", n, v.ListenAddress)
	return n.Version(), nil
}

func (n *NetNode) NewTransaction(
	ctx context.Context,
	t *proto.Transaction,
) (*proto.Transaction, error) {
	return &proto.Transaction{}, nil
}

func (n *NetNode) NewBlock(ctx context.Context, b *proto.Block) (*proto.Block, error) {
	return &proto.Block{}, nil
}

func (n *NetNode) Version() *proto.Version {
	return &proto.Version{
		Version:       "0.0.1",
		ListenAddress: n.ListenAddress,
		Peers:         n.Peers(),
	}
}

func (n *NetNode) addPeer(c client.Client, v *proto.Version) {
	if !n.canConnectWith(v.ListenAddress) {
		return
	}
	n.peers.addPeer(c, v)

	if len(v.Peers) > 0 {
		go func() {
			if err := n.BootstrapNetwork(v.Peers); err != nil {
				log.Fatalf("NetNode: %s, failed to bootstrap network: %v", n, err)
			}
		}()
	}
}

func (n *NetNode) dialRemote(address string) (client.Client, *proto.Version, error) {
	client, err := client.NewGRPCClient(address)
	if err != nil {
		return nil, nil, err
	}
	version, err := client.Handshake(context.Background(), n.Version())
	if err != nil {
		return nil, nil, err
	}
	n.logger.Infof("NetNode: %s, connected to %s", n, address)

	return client, version, nil
}

func (n *NetNode) BootstrapNetwork(addrs []string) error {
	time.Sleep(2 * time.Second) // TODO: handle endpoint not ready
	for _, addr := range addrs {
		if !n.canConnectWith(addr) {
			continue
		}
		client, version, err := n.dialRemote(addr)
		if err != nil {
			return err
		}
		n.addPeer(client, version)
	}
	return nil
}

func (n *NetNode) canConnectWith(addr string) bool {
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

func (n *NetNode) Peers() []string {
	return n.peers.List()
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
