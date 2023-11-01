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
	nodeAdrrs []string
}

func newNodeAPI() *nodeapi {
	return &nodeapi{
		nodeAdrrs: nodeAddrs,
	}
}

func (n *nodeapi) nodeApi() api.Client {
	return api.NewHTTPClient(n.nodeAdrrs[rand.Intn(nodesNum)])
}
