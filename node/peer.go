package node

import (
	"sync"

	"github.com/yuriykis/microblocknet/node/client"
	"github.com/yuriykis/microblocknet/proto"
)

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

type knownAddrs struct {
	addrs []string
	lock  sync.RWMutex
}

func newKnownAddrs() *knownAddrs {
	return &knownAddrs{
		addrs: make([]string, 0),
	}
}

func (ka *knownAddrs) append(addr string) {
	ka.lock.Lock()
	defer ka.lock.Unlock()
	ka.addrs = append(ka.addrs, addr)
}

func (ka *knownAddrs) update(addrs []string) {
	ka.lock.Lock()
	defer ka.lock.Unlock()
	ka.addrs = addrs
}

func (ka *knownAddrs) list() []string {
	ka.lock.Lock()
	defer ka.lock.Unlock()
	return ka.addrs
}
