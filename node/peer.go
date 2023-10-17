package node

import (
	"sync"
	"time"

	"github.com/yuriykis/microblocknet/node/client"
	"github.com/yuriykis/microblocknet/proto"
)

const checkConnectInterval = 15 * time.Second

type peer struct {
	*proto.Version
	lastPing time.Time
}

func newPeer(v *proto.Version) *peer {
	return &peer{
		Version:  v,
		lastPing: time.Now(),
	}
}

type peersMap struct {
	peers map[client.Client]*peer
	lock  sync.RWMutex
}

func NewPeersMap() *peersMap {
	return &peersMap{
		peers: make(map[client.Client]*peer),
	}
}
func (pm *peersMap) addPeer(c client.Client, v *proto.Version) {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	pm.peers[c] = newPeer(v)
}

func (pm *peersMap) removePeer(c client.Client) {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	delete(pm.peers, c)
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

func (pm *peersMap) listForPing() []string {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	peersList := make([]string, 0)
	for _, v := range pm.peers {
		if time.Since(v.lastPing) > checkConnectInterval {
			peersList = append(peersList, v.ListenAddress)
		}
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
