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

func (pm *peersMap) peersForPing() map[client.Client]*peer {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	peers := make(map[client.Client]*peer)
	for c, p := range pm.peers {
		if time.Since(p.lastPing) > checkConnectInterval {
			peers[c] = p
		}
	}
	return peers
}

func (pm *peersMap) updateLastPingTime(c client.Client) {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	pm.peers[c].lastPing = time.Now()
}

type knownAddrs struct {
	addrs map[string]int // [addr]pingAttempts
	lock  sync.RWMutex
}

func newKnownAddrs() *knownAddrs {
	return &knownAddrs{
		addrs: make(map[string]int),
	}
}

func (ka *knownAddrs) incPingAttempts(addr string) {
	ka.lock.Lock()
	defer ka.lock.Unlock()
	ka.addrs[addr]++
}

func (ka *knownAddrs) resetPingAttempts(addr string) {
	ka.lock.Lock()
	defer ka.lock.Unlock()
	ka.addrs[addr] = 0
}

func (ka *knownAddrs) append(addr string) {
	ka.lock.Lock()
	defer ka.lock.Unlock()
	ka.addrs[addr] = 0
}

func (ka *knownAddrs) update(addrs []string) {
	ka.lock.Lock()
	defer ka.lock.Unlock()
	for _, addr := range addrs {
		ka.addrs[addr] = 0
	}
}

func (ka *knownAddrs) list() []string {
	ka.lock.Lock()
	defer ka.lock.Unlock()
	addrsList := make([]string, 0)
	for addr := range ka.addrs {
		addrsList = append(addrsList, addr)
	}
	return addrsList
}
