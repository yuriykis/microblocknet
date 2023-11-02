package main

import (
	"math/rand"

	api "github.com/yuriykis/microblocknet/node/service/api_client"
)

const (
	nodesNum = 3
)

type nodeapi struct {
	peers      []api.Client
	knownHosts []string
}

func newNodeAPI() *nodeapi {
	n := &nodeapi{
		peers: make([]api.Client, 0),
		knownHosts: []string{
			"http://localhost:4000",
			"http://localhost:4001",
			"http://localhost:4002",
		},
	}
	n.makePeers()
	return n
}

func (n *nodeapi) NewHost(host string) {
	n.knownHosts = append(n.knownHosts, host)
	n.makePeers()
}

func (n *nodeapi) makePeers() {
	for _, addr := range n.knownHosts {
		n.peers = append(n.peers, api.NewHTTPClient(addr))
	}
	n.knownHosts = n.knownHosts[:0]
}

func (n *nodeapi) nodeApi() api.Client {
	return n.peers[rand.Intn(nodesNum)]
}
