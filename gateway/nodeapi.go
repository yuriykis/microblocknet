package main

import (
	"math/rand"

	api "github.com/yuriykis/microblocknet/node/service/api_client"
)

const (
	nodesNum = 3
)

var nodeAddrs = []string{"http://localhost:4000", "http://localhost:4001", "http://localhost:4002"}

type nodeapi struct {
	peers []api.Client
}

func newNodeAPI() *nodeapi {
	n := &nodeapi{}
	n.makePeers()
	return n
}

func (n *nodeapi) makePeers() {
	for _, addr := range nodeAddrs {
		n.peers = append(n.peers, api.NewHTTPClient(addr))
	}
}
func (n *nodeapi) nodeApi() api.Client {
	return n.peers[rand.Intn(nodesNum)]
}
