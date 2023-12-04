package service

import (
	"context"
	"fmt"
	"time"

	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/node/client"
	"go.uber.org/zap"
)

const (
	connectInterval    = 5 * time.Second
	pingInterval       = 6 * time.Second
	maxConnectAttempts = 100
)

// networkManager manages node's network connections
type networkManager struct {
	ListenAddress string
	peers         *peersMap
	knownAddrs    *knownAddrs
	logger        *zap.SugaredLogger

	quit
}

func NewNetworkManager(listenAddress string, logger *zap.SugaredLogger) *networkManager {
	return &networkManager{
		ListenAddress: listenAddress,
		peers:         NewPeersMap(),
		knownAddrs:    newKnownAddrs(),
		logger:        logger,
		quit: quit{
			tryConnectQuitCh: make(chan struct{}),
			pingQuitCh:       make(chan struct{}),
		},
	}
}

func (m *networkManager) String() string {
	return fmt.Sprintf("networkManager@%s", m.ListenAddress)
}

func (m *networkManager) start(bootstrapNodes []string) error {
	go m.tryConnect(m.tryConnectQuitCh, false)
	go m.ping(m.pingQuitCh, false)

	if len(bootstrapNodes) > 0 {
		go func() {
			if err := m.bootstrapNetwork(context.TODO(), bootstrapNodes); err != nil {
				m.logger.Errorf("node: %s, failed to bootstrap network: %v", m, err)
			}
		}()
	}

	return nil
}

func (m *networkManager) stop() error {
	m.shutdown()
	return nil
}

func (m *networkManager) bootstrapNetwork(ctx context.Context, addrs []string) error {
	for _, addr := range addrs {
		if !m.canConnectWith(addr) {
			continue
		}
		m.knownAddrs.append(addr, 0)
	}
	return nil
}

func (m *networkManager) peersAddrs(ctx context.Context) []string {
	return m.peers.Addresses()
}

func (m *networkManager) address() string {
	return m.ListenAddress
}

func (m *networkManager) dialRemote(address string) (client.Client, *proto.Version, error) {
	client, err := client.NewGRPCClient(address)
	if err != nil {
		return nil, nil, err
	}
	version, err := m.handshakeClient(client)
	if err != nil {
		return nil, nil, err
	}
	return client, version, nil
}

func (m *networkManager) version() *proto.Version {
	return &proto.Version{
		Version:       "0.0.1",
		ListenAddress: m.ListenAddress,
		Peers:         m.peersAddrs(context.TODO()),
	}
}

func (m *networkManager) Peers() map[client.Client]*peer {
	return m.peers.list()
}

func (m *networkManager) canConnectWith(addr string) bool {
	if addr == m.ListenAddress {
		return false
	}
	for _, peer := range m.peersAddrs(context.TODO()) {
		if peer == addr {
			return false
		}
	}
	return true
}

func (m *networkManager) handshakeClient(c client.Client) (*proto.Version, error) {
	version, err := c.Handshake(context.Background(), m.version())
	if err != nil {
		return nil, err
	}
	m.logger.Infof("node: %s, handshake with %s, version: %v", m, c, version)
	return version, nil
}

func (m *networkManager) addPeer(c client.Client, v *proto.Version) {
	if !m.canConnectWith(v.ListenAddress) {
		return
	}
	m.peers.addPeer(c, v)

	if len(v.Peers) > 0 {
		go func() {
			if err := m.bootstrapNetwork(context.TODO(), v.Peers); err != nil {
				m.logger.Errorf("node: %s, failed to bootstrap network: %v", m, err)
			}
		}()
	}
}

func (m *networkManager) broadcast(msg any) {
	for c := range m.Peers() {
		if err := m.sendMsg(c, msg); err != nil {
			m.logger.Errorf("node: %s, failed to send message to %s: %v", m, c, err)
		}
	}
}

func (m *networkManager) sendMsg(c client.Client, msg any) error {
	switch m := msg.(type) {
	case *proto.Transaction:
		_, err := c.NewTransaction(context.Background(), m)
		if err != nil {
			return err
		}
	case *proto.Block:
		_, err := c.NewBlock(context.Background(), m)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("node: %s, unknown message type: %v", m, msg)
	}
	return nil
}

// TryConnect tries to connect to known addresses
func (m *networkManager) tryConnect(quitCh chan struct{}, logging bool) {
	if logging {
		m.logger.Infof("node: %s, starting tryConnect\n", m)
	}
	for {
		select {
		case <-quitCh:
			if logging {
				m.logger.Infof("node: %s, stopping tryConnect\n", m)
			}
			return
		default:
			updatedKnownAddrs := make(map[string]int, 0)
			for addr, connectAttempts := range m.knownAddrs.list() {
				if !m.canConnectWith(addr) {
					continue
				}
				client, version, err := m.dialRemote(addr)
				if err != nil {
					errMsg := fmt.Sprintf(
						"node: %s, failed to connect to %s, will retry later: %v\n",
						m,
						addr,
						err)
					if connectAttempts >= maxConnectAttempts {
						errMsg = fmt.Sprintf(
							"node: %s, failed to connect to %s, reached maxConnectAttempts, removing from knownAddrs: %v\n",
							m,
							addr,
							err,
						)
					}
					if logging {
						m.logger.Errorf(errMsg)
					}
					// if the connection attemps is less than maxConnectAttempts, we will try to connect again
					// otherwise we will remove the address from the known addresses list
					// by not adding it to the updatedKnownAddrs map
					if connectAttempts < maxConnectAttempts {
						updatedKnownAddrs[addr] = connectAttempts + 1
					}
					continue
				}
				m.addPeer(client, version)
			}
			m.knownAddrs.update(updatedKnownAddrs)
			time.Sleep(connectInterval)
		}
	}
}

// Ping pings all known peers, if peer is not available,
// it will be removed from the peers list and added to the known addresses list
func (m *networkManager) ping(quitCh chan struct{}, logging bool) {
	if logging {
		m.logger.Infof("node: %s, starting ping\n", m)
	}
	for {
		select {
		case <-quitCh:
			if logging {
				m.logger.Infof("node: %s, stopping ping\n", m)
			}
			return
		default:
			for c, p := range m.peers.peersForPing() {
				_, err := m.handshakeClient(c)
				if err != nil {
					if logging {
						m.logger.Errorf("node: %s, failed to ping %s: %v\n", m, c, err)
					}
					m.knownAddrs.append(p.ListenAddress, 0)
					m.peers.removePeer(c)
					continue
				}
				m.peers.updateLastPingTime(c)
			}
			time.Sleep(pingInterval)
		}
	}
}

type quit struct {
	tryConnectQuitCh chan struct{}
	pingQuitCh       chan struct{}
}

func (q *quit) shutdown() {
	close(q.tryConnectQuitCh)
	close(q.pingQuitCh)
}
