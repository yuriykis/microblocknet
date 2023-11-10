package main

import (
	"context"
	"math/rand"

	api "github.com/yuriykis/microblocknet/node/service/api_client"
	"go.uber.org/zap"
)

type nodeapi struct {
	peers      []api.Client
	knownHosts []string
	logger     *zap.SugaredLogger
}

func newNodeAPI(logger *zap.SugaredLogger) *nodeapi {
	n := &nodeapi{
		peers:      make([]api.Client, 0),
		knownHosts: make([]string, 0),
		logger:     logger,
	}
	n.makePeers()
	return n
}

func (n *nodeapi) NewHost(host string) {
	for _, h := range n.knownHosts {
		if h == host {
			return
		}
	}
	n.knownHosts = append(n.knownHosts, host)
	n.makePeers()
}

func (n *nodeapi) RemovePeer(peer api.Client) {
	for i, p := range n.peers {
		if p.String() == peer.String() {
			n.peers = append(n.peers[:i], n.peers[i+1:]...)
			return
		}
	}
}

func (n *nodeapi) Peers() []api.Client {
	return n.peers
}

func (n *nodeapi) makePeers() {
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

func (n *nodeapi) pingPeer(peer api.Client) error {
	_, err := peer.Healthcheck(context.Background())
	if err != nil {
		n.RemovePeer(peer)
		return err
	}
	return nil
}

func (n *nodeapi) nodeApi() api.Client {
	if len(n.peers) == 0 {
		n.logger.Error("No peers available")
		return nil
	}
	peerSelected := n.selectPeer()
	n.logger.Infof("Selected peer: %s", n.peers[peerSelected])
	return n.peers[peerSelected]
}

func (n *nodeapi) selectPeer() int {
	return rand.Intn(len(n.peers))
}
