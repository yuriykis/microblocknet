package main

import (
	"fmt"
	"math/rand"

	api "github.com/yuriykis/microblocknet/node/service/api_client"
)

type nodeapi struct {
	peers      []api.Client
	knownHosts []string
}

func newNodeAPI() *nodeapi {
	n := &nodeapi{
		peers:      make([]api.Client, 0),
		knownHosts: make([]string, 0),
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

func (n *nodeapi) nodeApi() api.Client {
	if len(n.peers) == 0 {
		fmt.Println("No peers available")
		return nil
	}
	peerSelected := n.selectPeer()
	fmt.Println("Selected peer:", n.peers[peerSelected])
	return n.peers[peerSelected]
}

func (n *nodeapi) selectPeer() int {
	return rand.Intn(len(n.peers))
}
