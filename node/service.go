package node

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yuriykis/microblocknet/node/client"
	"github.com/yuriykis/microblocknet/proto"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	connectInterval = 1 * time.Second
	pingInterval    = 10 * time.Second
)

type Node interface {
	Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error)
	NewTransaction(ctx context.Context, t *proto.Transaction) (*proto.Transaction, error)
	NewBlock(ctx context.Context, b *proto.Block) (*proto.Block, error)
}

type NetNode struct {
	ListenAddress string

	logger     *zap.SugaredLogger
	peers      *peersMap
	knownAddrs *knownAddrs
}

func New(listenAddress string) *NetNode {
	return &NetNode{
		ListenAddress: listenAddress,
		peers:         NewPeersMap(),
		logger:        makeLogger(),
		knownAddrs:    newKnownAddrs(),
	}
}

func (n *NetNode) Start(listenAddr string, bootstrapNodes []string, server Server) error {
	go n.tryConnect()
	go n.ping()

	if len(bootstrapNodes) > 0 {
		go func() {
			if err := n.BootstrapNetwork(bootstrapNodes); err != nil {
				log.Fatalf("NetNode: %s, failed to bootstrap network: %v", n, err)
			}
		}()
	}
	return server.Serve()
}

func (n *NetNode) Stop(server Server) {
	server.Close()
}

func (n *NetNode) String() string {
	return n.ListenAddress
}

func (n *NetNode) Handshake(
	ctx context.Context,
	v *proto.Version,
) (*proto.Version, error) {
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
	version, err := n.handshakeClient(client)
	if err != nil {
		return nil, nil, err
	}
	return client, version, nil
}

func (n *NetNode) handshakeClient(c client.Client) (*proto.Version, error) {
	version, err := c.Handshake(context.Background(), n.Version())
	if err != nil {
		return nil, err
	}
	n.logger.Infof("NetNode: %s, handshake with %s, version: %v", n, c, version)
	return version, nil
}

func (n *NetNode) BootstrapNetwork(addrs []string) error {
	for _, addr := range addrs {
		if !n.canConnectWith(addr) {
			continue
		}
		n.knownAddrs.append(addr)
	}
	return nil
}

// TryConnect tries to connect to known addresses
func (n *NetNode) tryConnect() {
	for {
		updatedKnownAddrs := make([]string, 0)
		for _, addr := range n.knownAddrs.list() {
			if !n.canConnectWith(addr) {
				continue
			}
			client, version, err := n.dialRemote(addr)
			if err != nil {
				fmt.Printf(
					"NetNode: %s, failed to connect to %s, will retry later: %v\n",
					n,
					addr,
					err,
				)
				updatedKnownAddrs = append(updatedKnownAddrs, addr)
				n.knownAddrs.incPingAttempts(addr)
				continue
			}
			n.addPeer(client, version)
		}
		n.knownAddrs.update(updatedKnownAddrs)
		time.Sleep(connectInterval)
	}
}

// Ping pings all known peers, if peer is not available,
// it will be removed from the peers list and added to the known addresses list
func (n *NetNode) ping() {
	for {
		for c, p := range n.peers.peersForPing() {
			v, err := n.handshakeClient(c)
			if err != nil {
				fmt.Printf("NetNode: %s, failed to ping %s: %v\n", n, c, err)
				n.knownAddrs.append(p.ListenAddress)
				n.peers.removePeer(c)
				continue
			}
			n.knownAddrs.resetPingAttempts(v.ListenAddress)
			n.peers.updateLastPingTime(c)
		}
		time.Sleep(pingInterval)
	}
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
