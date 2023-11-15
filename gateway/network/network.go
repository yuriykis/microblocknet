package network

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	api "github.com/yuriykis/microblocknet/node/api_client"
	"go.uber.org/zap"
)

const pingPeersInterval = 5 * time.Second

type Networker interface {
	NewPeer(host string) error
	RemovePeer(peer api.Client) error
	Node() (api.Client, error)
}

type network struct {
	peers      []api.Client
	logger     *zap.SugaredLogger
	knownHosts []string
}

func New(logger *zap.SugaredLogger) Networker {
	n := &network{
		peers:      make([]api.Client, 0),
		knownHosts: make([]string, 0),
		logger:     logger,
	}
	n.makePeers()
	go n.pingKnownPeers()
	return n
}

func (n *network) NewPeer(host string) error {
	for _, h := range n.knownHosts {
		if h == host {
			return fmt.Errorf("host already known")
		}
	}
	n.knownHosts = append(n.knownHosts, host)
	n.makePeers()
	return nil
}

func (n *network) RemovePeer(peer api.Client) error {
	for i, p := range n.peers {
		if p.String() == peer.String() {
			n.peers = append(n.peers[:i], n.peers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("peer not found, nothing to remove")
}

func (n *network) Node() (api.Client, error) {
	if len(n.peers) == 0 {
		n.logger.Error("No peers available")
		return nil, fmt.Errorf("no peers available")
	}
	peerSelected := n.selectPeer()
	n.logger.Infof("Selected peer: %s", n.peers[peerSelected])
	return n.peers[peerSelected], nil
}

func (n *network) makePeers() {
	for _, addr := range n.knownHosts {
		n.peers = append(n.peers, api.NewHTTPClient(adjustAddr(addr)))
	}
	n.knownHosts = n.knownHosts[:0]
}

func adjustAddr(addr string) string {
	if len(addr) < 7 || addr[:7] != "http://" {
		addr = "http://" + addr
	}
	return addr
}

func (n *network) pingPeer(peer api.Client) error {
	_, err := peer.Healthcheck(context.Background())
	if err != nil {
		n.RemovePeer(peer)
		return err
	}
	return nil
}

func (n *network) selectPeer() int {
	return rand.Intn(len(n.peers))
}

func (n *network) peersApis() []api.Client {
	return n.peers
}

func (n *network) pingKnownPeers() {
	n.logger.Infof("starting to ping known peers")
	for {
		select {
		case <-time.After(pingPeersInterval):
			for _, addr := range n.peersApis() {
				if err := n.pingPeer(addr); err != nil {
					n.logger.Errorf("failed to ping peer: %v", err)
				}
			}
		}
	}
}
